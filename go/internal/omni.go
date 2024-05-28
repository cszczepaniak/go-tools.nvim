package internal

import (
	"github.com/cszczepaniak/go-tools/internal/constructor"
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/iferr"
)

func GenerateReplacements(
	filePath string,
	pos file.Position,
) (file.Replacement, error) {
	fns := []func(string, file.Position) (file.Replacement, error){
		constructor.Generate,
		iferr.Generate,
	}

	for _, fn := range fns {
		r, err := fn(filePath, pos)
		if err != nil {
			return file.Replacement{}, err
		}

		if len(r.Lines) != 0 {
			return r, nil
		}
	}

	return file.Replacement{}, nil
}
