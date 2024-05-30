package loader

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sync"

	"github.com/cszczepaniak/go-tools/internal/file"
	"golang.org/x/tools/go/packages"
)

type Loader struct {
	contents file.Contents

	fset *token.FileSet

	f     *ast.File
	fOnce sync.Once

	pkg     *packages.Package
	pkgOnce sync.Once
}

func New(
	contents file.Contents,
) Loader {
	return Loader{
		contents: contents,
		fset:     token.NewFileSet(),
	}
}

func (l *Loader) ParseFile() (*ast.File, error) {
	var err error
	l.fOnce.Do(func() {
		l.f, err = parser.ParseFile(
			l.fset,
			l.contents.AbsPath,
			l.contents.Contents,
			parser.AllErrors|parser.ParseComments,
		)
	})
	return l.f, err
}

// parseFile is used in loadPkg
func (l *Loader) parseFile(
	fset *token.FileSet,
	filepath string,
	src []byte,
) (*ast.File, error) {
	if filepath == l.contents.AbsPath {
		return l.ParseFile()
	}

	return parser.ParseFile(fset, filepath, src, parser.AllErrors|parser.ParseComments)
}

func (l *Loader) LoadPackage() error {
	var err error
	l.pkgOnce.Do(func() {
		dir, _ := filepath.Split(l.contents.AbsPath)

		var pkgs []*packages.Package
		pkgs, err = packages.Load(
			&packages.Config{
				Mode:      packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
				ParseFile: l.parseFile,
			},
			dir,
		)
		if err != nil {
			return
		}

		if len(pkgs) != 1 {
			panic(`should be unreachable; we only specify one package in the given pattern`)
		}

		l.pkg = pkgs[0]
	})
	return err
}
