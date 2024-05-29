package iferr

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
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
	}, dir)
	if err != nil {
		return file.Replacement{}, err
	}

	if len(pkgs) != 1 {
		return file.Replacement{}, errors.New("expected only one package")
	}

	pkg := pkgs[0]
	var astFile *ast.File

	for _, f := range pkg.Syntax {
		rng := asthelper.RangeFromNode(pkg.Fset, f)
		if rng.ContainsPos(pos) {
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
	astutil.Apply(astFile, func(c *astutil.Cursor) bool {
		_, ok := c.Node().(*ast.BlockStmt)
		if ok {
			indent++
			return true
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

		if len(assn.Lhs) == 0 {
			return false
		}

		if len(assn.Rhs) != 1 {
			return false
		}

		rhs := assn.Rhs[0]
		typ, ok := pkg.TypesInfo.Types[rhs]
		if !ok {
			fmt.Println(pkg.TypesInfo)
			return false
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
				return false
			}
		case *types.Named:
			t = typ
			idx = 0
		}

		if t == nil || t.Obj().Pkg() != nil || t.Obj().Name() != "error" {
			return false
		}

		if idx >= len(assn.Lhs) {
			return false
		}

		lhs := assn.Lhs[idx]
		ident, ok := lhs.(*ast.Ident)
		if !ok {
			return false
		}

		errName = ident.Name
		finalIndent = indent

		return false
	}, func(c *astutil.Cursor) bool {
		_, ok := c.Node().(*ast.BlockStmt)
		if ok {
			indent--
		}
		return true
	})

	if errName == "" {
		return file.Replacement{}, nil
	}

	var fnDecl *ast.FuncDecl
	for _, decl := range astFile.Decls {
		if !asthelper.RangeFromNode(pkg.Fset, decl).ContainsPos(pos) {
			continue
		}

		fd, ok := decl.(*ast.FuncDecl)
		if ok {
			fnDecl = fd
			break
		}
	}

	typ, ok := pkg.TypesInfo.Defs[fnDecl.Name]
	if !ok {
		return file.Replacement{}, errors.New("no type information for func decl")
	}

	sig, ok := typ.Type().(*types.Signature)
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
	fmt.Fprintf(w, "%sreturn ", strings.Repeat("\t", finalIndent+1))

	for i := 0; i < sig.Results().Len()-1; i++ {
		r := sig.Results().At(i)
		err := printZeroValue(w, r.Type())
		if err != nil {
			return file.Replacement{}, err
		}
		fmt.Fprint(w, ", ")
	}

	fmt.Fprint(w, errName)
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
