package asthelper

import (
	"go/ast"
	"go/token"

	"github.com/cszczepaniak/go-tools/internal"
)

func ClosestNodeOfType[T ast.Node](
	fset *token.FileSet,
	f *ast.File,
	pos internal.Position,
) T {
	var currNode T
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		tNode, ok := n.(T)
		if !ok {
			return true
		}

		rng := RangeFromNode(fset, n)
		if rng.ContainsPos(pos) {
			currNode = tNode
		}

		return true
	})

	return currNode
}

func RangeFromNode(fset *token.FileSet, n ast.Node) internal.Range {
	pStart := fset.PositionFor(n.Pos(), false)
	pEnd := fset.PositionFor(n.End(), false)

	return internal.Range{
		Start: internal.Position{
			Line: pStart.Line,
			Col:  pStart.Column,
		},
		Stop: internal.Position{
			Line: pEnd.Line,
			Col:  pEnd.Column,
		},
	}
}
