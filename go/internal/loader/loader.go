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

	Fset *token.FileSet

	fileOnce func() (*ast.File, error)
	pkgOnce  func() (*packages.Package, error)
}

func New(
	contents file.Contents,
) *Loader {
	fset := token.NewFileSet()

	l := &Loader{
		contents: contents,
		Fset:     fset,

		fileOnce: sync.OnceValues(func() (*ast.File, error) {
			return parser.ParseFile(
				fset,
				contents.AbsPath,
				contents.Contents,
				parser.AllErrors|parser.ParseComments,
			)
		}),
	}

	l.pkgOnce = sync.OnceValues(l.loadPackage)
	return l
}

func (l *Loader) ParseFile() (*ast.File, error) {
	return l.fileOnce()
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

	return parser.ParseFile(
		fset,
		filepath,
		src,
		parser.AllErrors|parser.ParseComments,
	)
}

func (l *Loader) LoadPackage() (*packages.Package, error) {
	return l.pkgOnce()
}

func (l *Loader) loadPackage() (*packages.Package, error) {
	dir, _ := filepath.Split(l.contents.AbsPath)

	pkgs, err := packages.Load(
		&packages.Config{
			Mode:      packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
			ParseFile: l.parseFile,
		},
		dir,
	)
	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		panic(`should be unreachable; we only specify one package in the given pattern`)
	}

	return pkgs[0], nil
}
