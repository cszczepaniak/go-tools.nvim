package selectorchain

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/cszczepaniak/go-tools/internal/asthelper"
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/linewriter"
	"github.com/cszczepaniak/go-tools/internal/logging"
	"github.com/cszczepaniak/go-tools/internal/suggestions"
)

func Generate(
	l suggestions.FileParser,
	contents file.Contents,
	offset int,
) (file.Replacement, error) {
	f, err := l.ParseFile()
	if err != nil {
		return file.Replacement{}, err
	}

	start := findStartOfChain(f.ASTPath)
	if start == nil {
		logging.Debug("selectorchain found no chain")
		return file.Replacement{}, nil
	}

	w := &linewriter.Writer{}
	err = formatChain(
		w,
		f.Fset.File(start.Pos()),
		f.IndentLevel(),
		contents,
		start,
	)
	if err != nil {
		return file.Replacement{}, err
	}

	rng := asthelper.RangeFromNode(f.Fset, start)

	return file.Replacement{
		Range: rng,
		Lines: w.TakeLines(),
	}, nil
}

func formatChain(
	w *linewriter.Writer,
	f *token.File,
	indent int,
	contents file.Contents,
	n ast.Node,
) error {
	switch n := n.(type) {
	case *ast.CallExpr:
		return formatCall(w, f, indent, contents, n)
	case *ast.SelectorExpr:
		return formatSel(w, f, indent, contents, n)
	case *ast.Ident:
		_, err := w.Write([]byte(n.Name))
		return err
	default:
		return fmt.Errorf("unexpected node type in chain: %T", n)
	}
}

func formatCall(
	w *linewriter.Writer,
	f *token.File,
	indent int,
	contents file.Contents,
	c *ast.CallExpr,
) error {
	err := formatChain(w, f, indent, contents, c.Fun)
	if err != nil {
		return err
	}

	start := f.Offset(c.Lparen)
	stop := f.Offset(c.Rparen) + 1
	_, err = w.Write(contents.BytesInRange(start, stop))
	return err
}

func formatSel(
	w *linewriter.Writer,
	f *token.File,
	indent int,
	contents file.Contents,
	s *ast.SelectorExpr,
) error {
	err := formatChain(w, f, indent, contents, s.X)
	if err != nil {
		return err
	}

	// Add one more indent because we're going to indent the calls that spill onto new lines.
	str := fmt.Sprintf(".\n%s", strings.Repeat("\t", indent+1))
	_, err = w.Write([]byte(str))
	if err != nil {
		return err
	}

	return formatChain(w, f, indent, contents, s.Sel)
}

func findStartOfChain(path []ast.Node) ast.Node {
	foundStart := false
	for i := 0; i < len(path)-1; i++ {
		curr := path[i]
		next := path[i+1]

		switch curr.(type) {
		case *ast.SelectorExpr:
			// If curr is a select, we keep going if next is a select or a call.
			switch next.(type) {
			case *ast.SelectorExpr, *ast.CallExpr:
				continue
			default:
				return curr
			}
		case *ast.CallExpr:
			// If curr is a call, we keep going only if next is a select
			switch next.(type) {
			case *ast.SelectorExpr:
				continue
			default:
				return curr
			}
		default:
			if !foundStart {
				continue
			}
			return nil
		}
	}

	return nil
}
