package loader

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sync"
	"sync/atomic"

	"github.com/cszczepaniak/go-tools/internal/asthelper"
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/logging"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type Loader struct {
	cursorOffset int

	Contents file.Contents

	fileOnce func() (File, error)
	pkgOnce  func() (*packages.Package, error)

	nFilesParsed        atomic.Int64
	whenWasMyFileParsed atomic.Int64
	nFunctionsStripped  atomic.Int64
	totalFunctionsSeen  atomic.Int64
}

func New(
	contents file.Contents,
	cursorOffset int,
) *Loader {
	l := &Loader{
		Contents:     contents,
		cursorOffset: cursorOffset,
	}

	l.fileOnce = sync.OnceValues(l.parseFile)
	l.pkgOnce = sync.OnceValues(l.loadPackage)
	return l
}

type File struct {
	// File is the parsed file.
	File *ast.File
	// Fset is the fileset that was used when parsing the file.
	Fset *token.FileSet
	// Pos is the cursor position translated to a token.Pos.
	Pos token.Pos
	// ASTPath is the path containing the node at Pos.
	ASTPath []ast.Node
}

func (f File) IndentLevel() int {
	indent := 0
	for i := len(f.ASTPath) - 1; i >= 0; i-- {
		if _, ok := f.ASTPath[i].(*ast.BlockStmt); ok {
			indent++
		}
	}
	return indent
}

func (l *Loader) ParseFile() (File, error) {
	return l.fileOnce()
}

func (l *Loader) parseFile() (File, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(
		fset,
		l.Contents.AbsPath,
		l.Contents.Contents,
		parser.AllErrors|parser.ParseComments,
	)
	if err != nil {
		return File{}, err
	}

	tokFile := fset.File(f.Pos())
	pos := tokFile.Pos(l.cursorOffset)
	astPath, _ := astutil.PathEnclosingInterval(f, pos, pos)

	return File{
		File:    f,
		Fset:    fset,
		ASTPath: astPath,
		Pos:     pos,
	}, nil
}

func (l *Loader) parseFileForLoadPkg(
	fset *token.FileSet,
	filepath string,
	src []byte,
) (*ast.File, error) {
	l.nFilesParsed.Add(1)
	var f *ast.File
	var pos token.Pos
	var err error
	if filepath == l.Contents.AbsPath {
		l.whenWasMyFileParsed.Store(l.nFilesParsed.Load())
		loadedFile, err := l.ParseFile()
		if err != nil {
			return nil, err
		}

		f = loadedFile.File
		pos = loadedFile.Pos
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
		// Strip away the bodies of all function declarations that don't contain our position. This
		// speeds up type checking.
		if fnDecl, ok := decl.(*ast.FuncDecl); ok {
			l.totalFunctionsSeen.Add(1)
			if asthelper.NodeContains(decl, pos) {
				continue
			}

			l.nFunctionsStripped.Add(1)
			fnDecl.Body = nil
		}
	}

	return f, nil
}

func (l *Loader) LoadPackage() (*packages.Package, error) {
	return l.pkgOnce()
}

func (l *Loader) loadPackage() (*packages.Package, error) {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode:      packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo,
			ParseFile: l.parseFileForLoadPkg,
		},
		fmt.Sprintf("file=%s", l.Contents.AbsPath),
	)
	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		panic(`should be unreachable; we only specify one package in the given pattern`)
	}

	logging.WithFields(map[string]any{
		"nFiles":    l.nFilesParsed.Load(),
		"myFileIdx": l.whenWasMyFileParsed.Load(),
		"nFuncs":    l.totalFunctionsSeen.Load(),
		"nStripped": l.nFunctionsStripped.Load(),
	}).Debug("package load stats")

	return pkgs[0], nil
}
