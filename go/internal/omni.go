package internal

import (
	"time"

	"github.com/cszczepaniak/go-tools/internal/constructor"
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/iferr"
	"github.com/cszczepaniak/go-tools/internal/loader"
	"github.com/cszczepaniak/go-tools/internal/logging"
)

func GenerateReplacements(
	contents file.Contents,
	offset int,
) (file.Replacement, error) {
	fns := map[string]func(*loader.Loader, int) (file.Replacement, error){
		"constructor": constructor.Generate,
		"iferr":       iferr.Generate,
	}

	order := []string{
		"constructor",
		"iferr",
	}

	l := loader.New(contents, offset)

	for _, name := range order {
		t0 := time.Now()
		r, err := fns[name](l, offset)
		logging.WithFields(map[string]any{"dur": time.Since(t0)}).Info(name + " finished")
		if err != nil {
			return file.Replacement{}, err
		}

		if len(r.Lines) != 0 {
			return r, nil
		}
	}

	return file.Replacement{}, nil
}
