package internal

import (
	"github.com/cszczepaniak/go-tools/internal/constructor"
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/iferr"
	"github.com/cszczepaniak/go-tools/internal/loader"
)

func GenerateReplacements(
	contents file.Contents,
	pos file.Position,
) (file.Replacement, error) {
	fns := []func(*loader.Loader, file.Position) (file.Replacement, error){
		constructor.Generate,
		iferr.Generate,
	}

	l := loader.New(contents, pos)

	for _, fn := range fns {
		r, err := fn(l, pos)
		if err != nil {
			return file.Replacement{}, err
		}

		if len(r.Lines) != 0 {
			return r, nil
		}
	}

	return file.Replacement{}, nil
}
