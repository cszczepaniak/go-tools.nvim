package constructor

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"strings"
	"unicode"

	"github.com/cszczepaniak/go-tools/internal/asthelper"
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/linewriter"
	"github.com/cszczepaniak/go-tools/internal/loader"
)

func Generate(
	l *loader.Loader,
	offset int,
) (file.Replacement, error) {
	pkg, err := l.LoadPackage()
	if err != nil {
		return file.Replacement{}, err
	}

	var typeDecl *ast.GenDecl
	var typeSpec *ast.TypeSpec

	for i, n := range l.ASTPath {
		if ts, ok := n.(*ast.TypeSpec); ok {
			typeSpec = ts
			if i+1 < len(l.ASTPath) {
				parent := l.ASTPath[i+1]
				if decl, ok := parent.(*ast.GenDecl); ok {
					typeDecl = decl
				}
			}
			break
		}
	}

	if typeDecl == nil || typeSpec == nil || typeSpec.Name == nil {
		return file.Replacement{}, nil
	}

	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return file.Replacement{}, nil
	}

	def, ok := pkg.TypesInfo.Defs[typeSpec.Name]
	if !ok {
		return file.Replacement{}, errors.New("no type info for struct type")
	}

	typ := def.Type()

	var t *types.Struct
	for typ != nil {
		if st, ok := typ.(*types.Struct); ok {
			t = st
			break
		}

		typ = typ.Underlying()
	}

	lw := &linewriter.Writer{}

	err = format.Node(lw, l.Fset, typeDecl)
	if err != nil {
		return file.Replacement{}, err
	}

	lw.Flush()

	// Add a blank line between.
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

	type fieldInfo struct {
		typeStr      string
		nameInStruct string
		nameInFunc   string
	}

	idx := -1
	maxStructMemberLen := 0
	var fields []fieldInfo
	for _, fld := range structType.Fields.List {
		typStr, err := formatNodeToString(fld.Type)
		if err != nil {
			return file.Replacement{}, err
		}

		if len(fld.Names) == 0 {
			idx++
			v := t.Field(idx)

			fields = append(fields, fieldInfo{
				typeStr:      typStr,
				nameInStruct: v.Name(),
				nameInFunc:   lowerFirstRune(typStr),
			})
			maxStructMemberLen = max(maxStructMemberLen, len(v.Name()))
		} else {
			for _, n := range fld.Names {
				idx++

				fields = append(fields, fieldInfo{
					typeStr:      typStr,
					nameInStruct: n.Name,
					nameInFunc:   lowerFirstRune(n.Name),
				})
				maxStructMemberLen = max(maxStructMemberLen, len(n.Name))
			}
		}
	}

	for _, f := range fields {
		lw.WriteLinef("\t%s %s,", f.nameInFunc, f.typeStr)
	}

	lw.WriteLinef(") %s {", typeSpec.Name.Name)
	lw.WriteLinef("\treturn %s{", typeSpec.Name.Name)

	for _, f := range fields {
		padding := maxStructMemberLen - len(f.nameInStruct)
		lw.WriteLinef("\t\t%s: %s%s,", f.nameInStruct, strings.Repeat(" ", padding), f.nameInFunc)
	}

	lw.WriteLinef("\t}")
	lw.WriteLinef("}")

	return file.Replacement{
		Range: asthelper.RangeFromNode(l.Fset, typeDecl),
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
