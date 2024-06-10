package iferr

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"strings"

	"github.com/cszczepaniak/go-tools/internal/asthelper"
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/linewriter"
	"github.com/cszczepaniak/go-tools/internal/loader"
	"github.com/cszczepaniak/go-tools/internal/logging"
)

func Generate(
	l *loader.Loader,
	offset int,
) (file.Replacement, error) {
	e := logging.WithFields(map[string]any{"handler": "iferr"})

	f, err := l.ParseFile()
	if err != nil {
		return file.Replacement{}, err
	}

	assnStmt, surrounding := findAssignmentAndSurroundingFunc(f.ASTPath)
	if surrounding == nil || assnStmt == nil {
		e.WithFields(map[string]any{
			"surroundingNil": surrounding == nil,
			"assnNil":        assnStmt == nil,
		}).Info("surrounding function or assignment statement not found")
		return file.Replacement{}, nil
	}

	replacementRange := asthelper.RangeFromNode(f.Fset, assnStmt)
	finalIndent := f.IndentLevel()

	pkg, err := l.LoadPackage()
	if err != nil {
		return file.Replacement{}, err
	}

	var funcTyp types.Type
	switch s := surrounding.(type) {
	case *ast.FuncDecl:
		funcTyp = pkg.TypesInfo.TypeOf(s.Name)
	case *ast.FuncLit:
		funcTyp = pkg.TypesInfo.TypeOf(s)
	}

	if funcTyp == nil {
		return file.Replacement{}, errors.New("type info not found for surrounding function")
	}

	errName := ""
	for _, e := range assnStmt.Lhs {
		if id, ok := e.(*ast.Ident); ok {
			t := pkg.TypesInfo.ObjectOf(id)
			if t != nil && isErrorType(t.Type()) {
				errName = id.Name
			}
		}
	}

	if errName == "" {
		e.Info("lhs did not have an error type")
		return file.Replacement{}, nil
	}

	sig, ok := funcTyp.(*types.Signature)
	if !ok {
		return file.Replacement{}, errors.New("not a signature")
	}

	w := &linewriter.Writer{}

	tokFile := f.Fset.File(assnStmt.Pos())
	start := tokFile.Offset(assnStmt.Pos())
	stop := tokFile.Offset(assnStmt.End())
	bs := l.Contents.BytesInRange(start, stop)

	w.Write(bs)
	w.Flush()

	// err = format.Node(w, pkg.Fset, assnStmt)
	// if err != nil {
	// 	return file.Replacement{}, err
	// }
	//
	// logging.Debug("done formatting node")
	// w.Flush()
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
				err := printZeroValue(w, pkg.PkgPath, r.Type())
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

func findAssignmentAndSurroundingFunc(
	path []ast.Node,
) (*ast.AssignStmt, ast.Node) {
	var assnStmt *ast.AssignStmt
	for _, n := range path {
		switch n := n.(type) {
		case *ast.AssignStmt:
			if assnStmt == nil {
				assnStmt = n
			}
		case *ast.FuncDecl:
			if assnStmt != nil {
				return assnStmt, n
			}
		case *ast.FuncLit:
			if assnStmt != nil {
				return assnStmt, n
			}
		}
	}
	return nil, nil
}

func printZeroValue(w io.Writer, myPkg string, typ types.Type) error {
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
			if tr.Obj().Pkg() != nil && tr.Obj().Pkg().Path() != myPkg {
				fmt.Fprintf(w, "%s.", tr.Obj().Pkg().Name())
			}
			fmt.Fprintf(w, "%s{}", tr.Obj().Name())
		} else {
			return printZeroValue(w, myPkg, tr.Underlying())
		}
	case *types.Map, *types.Slice, *types.Interface, *types.Pointer:
		fmt.Fprint(w, "nil")
	case *types.Array:
		fmt.Fprintf(w, "%s{}", tr.String())
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
