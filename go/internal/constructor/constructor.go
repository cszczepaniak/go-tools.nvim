package constructor

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"unicode"

	"github.com/cszczepaniak/go-tools/internal"
	"github.com/cszczepaniak/go-tools/internal/asthelper"
)

func Generate(
	filePath string,
	pos internal.Position,
) (internal.Replacement, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		return internal.Replacement{}, err
	}

	var typeDecl *ast.GenDecl
	var typeSpec *ast.TypeSpec
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}

		rng := asthelper.RangeFromNode(fset, decl)
		if rng.ContainsPos(pos) {
			typeDecl = gd

			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				rng = asthelper.RangeFromNode(fset, spec)
				if rng.ContainsPos(pos) {
					typeSpec = ts
					break
				}
			}

			break
		}
	}

	if typeSpec == nil || typeSpec.Name == nil {
		return internal.Replacement{}, nil
	}

	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return internal.Replacement{}, nil
	}

	lw := &internal.LineWriter{}

	err = format.Node(lw, fset, typeDecl)
	if err != nil {
		return internal.Replacement{}, err
	}

	lw.Flush()

	// Add a blank lines between.
	lw.WriteLinef("")

	var fnName string
	structName := []rune(typeSpec.Name.Name)
	if unicode.IsLower(structName[0]) {
		structName[0] = unicode.ToUpper(structName[0])
		fnName = "new" + string(structName)
	} else {
		fnName = "New" + typeSpec.Name.Name
	}

	lw.WriteLinef("func %s(", fnName)

	for _, fld := range structType.Fields.List {
		typStr, err := formatNodeToString(fld.Type)
		if err != nil {
			return internal.Replacement{}, err
		}
		for _, n := range fld.Names {
			lw.WriteLinef("\t%s %s,", lowerFirstRune(n.Name), typStr)
		}
	}

	lw.WriteLinef(") %s {", typeSpec.Name.Name)
	lw.WriteLinef("\treturn %s{", typeSpec.Name.Name)

	for _, fld := range structType.Fields.List {
		for _, n := range fld.Names {
			lw.WriteLinef("\t\t%s: %s,", n.Name, lowerFirstRune(n.Name))
		}
	}

	lw.WriteLinef("\t}")
	lw.WriteLinef("}")

	return internal.Replacement{
		Range: asthelper.RangeFromNode(fset, typeDecl),
		Lines: lw.TakeLines(),
	}, err
}

func lowerFirstRune(str string) string {
	rs := []rune(str)
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}

func formatNodeToString(n ast.Node) (string, error) {
	sb := &strings.Builder{}
	err := format.Node(sb, token.NewFileSet(), n)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

func getASTFileAndPosition(
	fName string,
	fPos int,
	fset *token.FileSet,
	syntax []*ast.File,
) (*ast.File, token.Pos, error) {
	var translatedPos token.Pos = -1
	fset.Iterate(func(f *token.File) bool {
		if f.Name() == fName {
			translatedPos = f.Pos(fPos)
			return false
		}

		return true
	})

	var theFile *ast.File
	for _, f := range syntax {
		fmt.Println(f.Name)
		if f.Pos() <= translatedPos && translatedPos <= f.End() {
			theFile = f
			break
		}
	}

	if translatedPos == -1 {
		return nil, 0, errors.New("did not find position in the given file")
	}

	if theFile == nil {
		return nil, 0, errors.New("did not find AST for file")
	}

	return theFile, translatedPos, nil
}
