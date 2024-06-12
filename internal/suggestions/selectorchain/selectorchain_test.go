package selectorchain

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/linewriter"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
	"golang.org/x/tools/go/ast/astutil"
)

func TestFindStartOfChain_AllCalls(t *testing.T) {
	src := `package foo

	func foo() {
	A().B().C().D()
}`

	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	must.NoError(t, err)

	a := findIdent(t, root, "A")
	f := fset.File(a.Pos())

	path, _ := astutil.PathEnclosingInterval(root, a.Pos(), a.Pos())
	startA := findStartOfChain(path)

	call, ok := startA.(*ast.CallExpr)
	must.True(t, ok)

	sel, ok := call.Fun.(*ast.SelectorExpr)
	must.True(t, ok)

	test.Eq(t, "D", sel.Sel.Name)

	b := findIdent(t, root, "B")
	path, _ = astutil.PathEnclosingInterval(root, b.Pos(), b.Pos())
	startB := findStartOfChain(path)
	test.Eq(t, startA, startB)

	c := findIdent(t, root, "C")
	path, _ = astutil.PathEnclosingInterval(root, c.Pos(), c.Pos())
	startC := findStartOfChain(path)
	test.Eq(t, startA, startC)

	d := findIdent(t, root, "D")
	path, _ = astutil.PathEnclosingInterval(root, d.Pos(), d.Pos())
	startD := findStartOfChain(path)
	test.Eq(t, startA, startD)

	w := &linewriter.Writer{}
	err = formatChain(
		w,
		f,
		0,
		file.Contents{
			Contents: []byte(src),
		},
		startA,
	)
	must.NoError(t, err)
	test.Eq(t, []string{"A().", "\tB().", "\tC().", "\tD()"}, w.TakeLines())
}

func TestFindStartOfChain_AllCalls_ButWithParams(t *testing.T) {
	src := `package foo

	func foo() {
	A().B(x, y).C(y, z).D()
}`

	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	must.NoError(t, err)

	a := findIdent(t, root, "A")
	f := fset.File(a.Pos())

	path, _ := astutil.PathEnclosingInterval(root, a.Pos(), a.Pos())
	startA := findStartOfChain(path)

	call, ok := startA.(*ast.CallExpr)
	must.True(t, ok)

	sel, ok := call.Fun.(*ast.SelectorExpr)
	must.True(t, ok)

	test.Eq(t, "D", sel.Sel.Name)

	b := findIdent(t, root, "B")
	path, _ = astutil.PathEnclosingInterval(root, b.Pos(), b.Pos())
	startB := findStartOfChain(path)
	test.Eq(t, startA, startB)

	c := findIdent(t, root, "C")
	path, _ = astutil.PathEnclosingInterval(root, c.Pos(), c.Pos())
	startC := findStartOfChain(path)
	test.Eq(t, startA, startC)

	d := findIdent(t, root, "D")
	path, _ = astutil.PathEnclosingInterval(root, d.Pos(), d.Pos())
	startD := findStartOfChain(path)
	test.Eq(t, startA, startD)

	w := &linewriter.Writer{}
	err = formatChain(
		w,
		f,
		0,
		file.Contents{
			Contents: []byte(src),
		},
		startA,
	)
	must.NoError(t, err)
	test.Eq(t, []string{"A().", "\tB(x, y).", "\tC(y, z).", "\tD()"}, w.TakeLines())
}

func TestFindStartOfChain_AllCalls_ButWithParams_AndNewlines(t *testing.T) {
	src := `package foo

	func foo() {
	A().B(
		x,
		y,
	).C(y, z).D()
}`

	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	must.NoError(t, err)

	a := findIdent(t, root, "A")
	f := fset.File(a.Pos())

	path, _ := astutil.PathEnclosingInterval(root, a.Pos(), a.Pos())
	startA := findStartOfChain(path)

	call, ok := startA.(*ast.CallExpr)
	must.True(t, ok)

	sel, ok := call.Fun.(*ast.SelectorExpr)
	must.True(t, ok)

	test.Eq(t, "D", sel.Sel.Name)

	b := findIdent(t, root, "B")
	path, _ = astutil.PathEnclosingInterval(root, b.Pos(), b.Pos())
	startB := findStartOfChain(path)
	test.Eq(t, startA, startB)

	c := findIdent(t, root, "C")
	path, _ = astutil.PathEnclosingInterval(root, c.Pos(), c.Pos())
	startC := findStartOfChain(path)
	test.Eq(t, startA, startC)

	d := findIdent(t, root, "D")
	path, _ = astutil.PathEnclosingInterval(root, d.Pos(), d.Pos())
	startD := findStartOfChain(path)
	test.Eq(t, startA, startD)

	w := &linewriter.Writer{}
	err = formatChain(
		w,
		f,
		0,
		file.Contents{
			Contents: []byte(src),
		},
		startA,
	)
	must.NoError(t, err)
	test.Eq(
		t,
		[]string{
			"A().",
			"\tB(",
			"\t\tx,",
			"\t\ty,",
			"\t).",
			"\tC(y, z).",
			"\tD()",
		},
		w.TakeLines(),
	)
}

func TestFindStartOfChain_AllNonCalls(t *testing.T) {
	src := `package foo

	func foo() {
	A.B.C.D
}`

	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	must.NoError(t, err)

	a := findIdent(t, root, "A")
	f := fset.File(a.Pos())

	path, _ := astutil.PathEnclosingInterval(root, a.Pos(), a.Pos())
	startA := findStartOfChain(path)

	sel, ok := startA.(*ast.SelectorExpr)
	must.True(t, ok)

	test.Eq(t, "D", sel.Sel.Name)

	b := findIdent(t, root, "B")
	path, _ = astutil.PathEnclosingInterval(root, b.Pos(), b.Pos())
	startB := findStartOfChain(path)
	test.Eq(t, startA, startB)

	c := findIdent(t, root, "C")
	path, _ = astutil.PathEnclosingInterval(root, c.Pos(), c.Pos())
	startC := findStartOfChain(path)
	test.Eq(t, startA, startC)

	d := findIdent(t, root, "D")
	path, _ = astutil.PathEnclosingInterval(root, d.Pos(), d.Pos())
	startD := findStartOfChain(path)
	test.Eq(t, startA, startD)

	w := &linewriter.Writer{}
	err = formatChain(
		w,
		f,
		0,
		file.Contents{
			Contents: []byte(src),
		},
		startA,
	)
	must.NoError(t, err)
	test.Eq(t, []string{"A.", "\tB.", "\tC.", "\tD"}, w.TakeLines())
}

func TestFindStartOfChain_MixedCallsAndNonCalls(t *testing.T) {
	src := `package foo

	func foo() {
	A.B().C.D()
}`

	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	must.NoError(t, err)

	a := findIdent(t, root, "A")

	path, _ := astutil.PathEnclosingInterval(root, a.Pos(), a.Pos())
	startA := findStartOfChain(path)

	call, ok := startA.(*ast.CallExpr)
	must.True(t, ok)

	sel, ok := call.Fun.(*ast.SelectorExpr)
	must.True(t, ok)

	test.Eq(t, "D", sel.Sel.Name)

	b := findIdent(t, root, "B")
	path, _ = astutil.PathEnclosingInterval(root, b.Pos(), b.Pos())
	startB := findStartOfChain(path)
	test.Eq(t, startA, startB)

	c := findIdent(t, root, "C")
	path, _ = astutil.PathEnclosingInterval(root, c.Pos(), c.Pos())
	startC := findStartOfChain(path)
	test.Eq(t, startA, startC)

	d := findIdent(t, root, "D")
	path, _ = astutil.PathEnclosingInterval(root, d.Pos(), d.Pos())
	startD := findStartOfChain(path)
	test.Eq(t, startA, startD)
}

func findIdent(t *testing.T, root *ast.File, name string) *ast.Ident {
	t.Helper()

	var id *ast.Ident
	ast.Inspect(root, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == name {
			id = ident
			return false
		}

		return true
	})

	must.NotNil(t, id)
	return id
}
