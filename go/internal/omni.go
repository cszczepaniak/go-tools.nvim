package internal

import (
	"github.com/cszczepaniak/go-tools/internal/constructor"
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/iferr"
)

func GenerateReplacements(
	contents file.Contents,
	pos file.Position,
) (file.Replacement, error) {
	fns := []func(file.Contents, file.Position) (file.Replacement, error){
		constructor.Generate,
		iferr.Generate,
	}

	for _, fn := range fns {
		r, err := fn(contents, pos)
		if err != nil {
			return file.Replacement{}, err
		}

		if len(r.Lines) != 0 {
			return r, nil
		}
	}

	return file.Replacement{}, nil
}
