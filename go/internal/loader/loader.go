package loader

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sync"

	"github.com/cszczepaniak/go-tools/internal/asthelper"
	"github.com/cszczepaniak/go-tools/internal/file"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type Loader struct {
	contents     file.Contents
	cursorOffset int

	Fset    *token.FileSet
	Pos     token.Pos
	ASTPath []ast.Node

	fileOnce func() (*ast.File, error)
	pkgOnce  func() (*packages.Package, error)
}

func New(
	contents file.Contents,
	cursorOffset int,
) *Loader {
	fset := token.NewFileSet()

	l := &Loader{
		contents:     contents,
		cursorOffset: cursorOffset,

		Fset: fset,
	}

	l.fileOnce = sync.OnceValues(l.parseFile)
	l.pkgOnce = sync.OnceValues(l.loadPackage)
	return l
}

func (l *Loader) ParseFile() (*ast.File, error) {
	return l.fileOnce()
}

func (l *Loader) parseFile() (*ast.File, error) {
	f, err := parser.ParseFile(
		l.Fset,
		l.contents.AbsPath,
		l.contents.Contents,
		parser.AllErrors|parser.ParseComments,
	)
	if err != nil {
		return nil, err
	}

	tokFile := l.Fset.File(f.Pos())
	l.Pos = token.Pos(tokFile.Base() + l.cursorOffset)
	l.ASTPath, _ = astutil.PathEnclosingInterval(f, l.Pos, l.Pos)

	return f, nil
}

func (l *Loader) parseFileForLoadPkg(
	fset *token.FileSet,
	filepath string,
	src []byte,
) (*ast.File, error) {
	var f *ast.File
	var err error
	if filepath == l.contents.AbsPath {
		f, err = l.ParseFile()
	} else {
		f, err = parser.ParseFile(
			fset,
			filepath,
			src,
			parser.AllErrors|parser.ParseComments,
		)
	}
	if err != nil {
		return nil, err
	}

	for _, decl := range f.Decls {
		if asthelper.NodeContains(decl, l.Pos) {
			continue
		}

		// Strip away the bodies of all function declarations that don't contain our position. This
		// speeds up type checking.
		if fnDecl, ok := decl.(*ast.FuncDecl); ok {
			fnDecl.Body = nil
		}
	}

	return f, nil
}

func (l *Loader) LoadPackage() (*packages.Package, error) {
	return l.pkgOnce()
}

func (l *Loader) loadPackage() (*packages.Package, error) {
	dir, _ := filepath.Split(l.contents.AbsPath)

	pkgs, err := packages.Load(
		&packages.Config{
			Mode:      packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
			ParseFile: l.parseFileForLoadPkg,
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
