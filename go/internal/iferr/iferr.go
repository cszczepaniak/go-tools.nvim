package iferr

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"path/filepath"
	"strings"

	"github.com/cszczepaniak/go-tools/internal/asthelper"
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/linewriter"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func Generate(
	contents file.Contents,
	pos file.Position,
) (file.Replacement, error) {
	dir, _ := filepath.Split(contents.AbsPath)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			if filename == contents.AbsPath {
				src = contents.Contents
			}
			return parser.ParseFile(fset, filename, src, parser.AllErrors|parser.ParseComments)
		},
	}, dir)
	if err != nil {
		return file.Replacement{}, err
	}

	if len(pkgs) != 1 {
		return file.Replacement{}, errors.New("expected only one package")
	}

	pkg := pkgs[0]
	var astFile *ast.File

	nodeContainsPos := func(n ast.Node) bool {
		return asthelper.RangeFromNode(pkg.Fset, n).ContainsPos(pos)
	}

	for _, f := range pkg.Syntax {
		if nodeContainsPos(f) {
			astFile = f
			break
		}
	}

	if astFile == nil {
		return file.Replacement{}, errors.New("did not find syntax tree for position")
	}

	indent := 0
	errName := ""
	var replacementRange file.Range
	var assnStmt *ast.AssignStmt
	var finalIndent int
	var surrounding ast.Node
	astutil.Apply(astFile, func(c *astutil.Cursor) bool {
		_, ok := c.Node().(*ast.BlockStmt)
		if ok {
			indent++
			return true
		}

		fd, ok := c.Node().(*ast.FuncDecl)
		if ok && nodeContainsPos(fd) {
			surrounding = fd
		}

		fl, ok := c.Node().(*ast.FuncLit)
		if ok && nodeContainsPos(fl) {
			surrounding = fl
		}

		assn, ok := c.Node().(*ast.AssignStmt)
		if !ok {
			return true
		}

		rng := asthelper.RangeFromNode(pkg.Fset, assn)
		if !rng.ContainsPos(pos) {
			return true
		}
		replacementRange = rng
		assnStmt = assn
		finalIndent = indent

		return true
	}, func(c *astutil.Cursor) bool {
		_, ok := c.Node().(*ast.BlockStmt)
		if ok {
			indent--
		}
		return true
	})

	if surrounding == nil || assnStmt == nil {
		return file.Replacement{}, nil
	}

	var funcTyp types.Type
	switch s := surrounding.(type) {
	case *ast.FuncDecl:
		t, ok := pkg.TypesInfo.Defs[s.Name]
		if !ok {
			return file.Replacement{}, errors.New("type info not found for func declaration")
		}
		funcTyp = t.Type()
	case *ast.FuncLit:
		t, ok := pkg.TypesInfo.Types[s]
		if !ok {
			return file.Replacement{}, errors.New("type info not found for func literal")
		}
		funcTyp = t.Type
	default:
		return file.Replacement{}, fmt.Errorf("dev error: unexpected surrounding %T", surrounding)
	}

	if len(assnStmt.Lhs) == 0 {
		return file.Replacement{}, nil
	}

	if len(assnStmt.Rhs) != 1 {
		return file.Replacement{}, nil
	}

	rhs := assnStmt.Rhs[0]
	typ, ok := pkg.TypesInfo.Types[rhs]
	if !ok {
		return file.Replacement{}, nil
	}

	var t *types.Named
	var idx int
	switch typ := typ.Type.(type) {
	case *types.Tuple:
		val := typ.At(typ.Len() - 1).Type()
		if tt, ok := val.(*types.Named); ok {
			t = tt
			idx = typ.Len() - 1
		} else {
			return file.Replacement{}, nil
		}
	case *types.Named:
		t = typ
		idx = 0
	}

	if !isErrorType(t) {
		return file.Replacement{}, nil
	}

	if idx >= len(assnStmt.Lhs) {
		return file.Replacement{}, nil
	}

	lhs := assnStmt.Lhs[idx]
	ident, ok := lhs.(*ast.Ident)
	if !ok {
		return file.Replacement{}, nil
	}

	errName = ident.Name

	if errName == "" {
		return file.Replacement{}, nil
	}

	sig, ok := funcTyp.(*types.Signature)
	if !ok {
		return file.Replacement{}, errors.New("not a signature")
	}

	w := &linewriter.Writer{}

	err = format.Node(w, pkg.Fset, assnStmt)
	if err != nil {
		return file.Replacement{}, err
	}

	w.Flush()
	w.WriteLinef("%sif %s != nil {", strings.Repeat("\t", finalIndent), errName)

	totalResults := sig.Results().Len()

	errIdx := -1
	for i := 0; i < totalResults; i++ {
		v := sig.Results().At(i)
		if isErrorType(v.Type()) {
			errIdx = i
			break
		}
	}

	fmt.Fprint(w, strings.Repeat("\t", finalIndent+1))
	if totalResults == 0 || errIdx == -1 {
		// If the function we're in does not return anything or doesn't return an error
		// anywhere, just panic with the error.
		fmt.Fprintf(w, "panic(%s)", errName)
	} else {
		fmt.Fprint(w, "return ")

		for i := 0; i < totalResults; i++ {
			if i == errIdx {
				fmt.Fprint(w, errName)
			} else {
				r := sig.Results().At(i)
				err := printZeroValue(w, r.Type())
				if err != nil {
					return file.Replacement{}, err
				}
			}

			if i < totalResults-1 {
				fmt.Fprint(w, ", ")
			}
		}
	}
	w.Flush()
	w.WriteLinef("%s}", strings.Repeat("\t", finalIndent))

	return file.Replacement{
		Range: replacementRange,
		Lines: w.TakeLines(),
	}, nil
}

func printZeroValue(w io.Writer, typ types.Type) error {
	switch tr := typ.(type) {
	case *types.Basic:
		switch {
		case tr.Info()&types.IsBoolean > 0:
			fmt.Fprint(w, "false")
		case tr.Info()&types.IsNumeric > 0:
			fmt.Fprint(w, "0")
		case tr.Info()&types.IsString > 0:
			fmt.Fprint(w, `""`)
		}
	case *types.Named:
		if _, ok := tr.Underlying().(*types.Struct); ok {
			if tr.Obj().Pkg() != nil {
				fmt.Fprintf(w, "%s.", tr.Obj().Pkg().Name())
			}
			fmt.Fprintf(w, "%s{}", tr.Obj().Name())
		} else {
			return printZeroValue(w, tr.Underlying())
		}
	case *types.Map, *types.Array, *types.Interface, *types.Pointer:
		fmt.Fprint(w, "nil")
	default:
		return fmt.Errorf("I don't know how to handle %T", typ)
	}

	return nil
}

func isErrorType(typ types.Type) bool {
	n, ok := typ.(*types.Named)
	if !ok {
		return false
	}

	return n != nil && n.Obj().Pkg() == nil && n.Obj().Name() == "error"
}
