package constructor

import (
	"errors"
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
	f, err := l.ParseFile()
	if err != nil {
		return file.Replacement{}, err
	}

	var typeDecl *ast.GenDecl
	var typeSpec *ast.TypeSpec

	for i, n := range f.ASTPath {
		if ts, ok := n.(*ast.TypeSpec); ok {
			typeSpec = ts
			if i+1 < len(f.ASTPath) {
				parent := f.ASTPath[i+1]
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

	lw := &linewriter.Writer{}

	tokFile := f.Fset.File(typeDecl.Pos())
	start := tokFile.Offset(typeDecl.Pos())
	stop := tokFile.Offset(typeDecl.End())
	bs := l.Contents.BytesInRange(start, stop)

	lw.Write(bs)
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

			t, err := loadStructType(l, typeSpec)
			if err != nil {
				return file.Replacement{}, err
			}

			v := t.Field(idx)

			fields = append(fields, fieldInfo{
				typeStr:      typStr,
				nameInStruct: v.Name(),
				nameInFunc:   lowerFirstRune(v.Name()),
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
		Range: asthelper.RangeFromNode(f.Fset, typeDecl),
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

func loadStructType(l *loader.Loader, typeSpec *ast.TypeSpec) (*types.Struct, error) {
	pkg, err := l.LoadPackage()
	if err != nil {
		return nil, err
	}

	def, ok := pkg.TypesInfo.Defs[typeSpec.Name]
	if !ok {
		return nil, errors.New("no type info for struct type")
	}

	typ := def.Type()

	for typ != nil {
		if st, ok := typ.(*types.Struct); ok {
			return st, nil
		}

		typ = typ.Underlying()
	}

	return nil, nil
}
