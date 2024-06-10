package internal

import (
	"time"

	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/loader"
	"github.com/cszczepaniak/go-tools/internal/logging"
	"github.com/cszczepaniak/go-tools/internal/suggestions"
	"github.com/cszczepaniak/go-tools/internal/suggestions/constructor"
	"github.com/cszczepaniak/go-tools/internal/suggestions/iferr"
)

func GenerateReplacements(
	contents file.Contents,
	offset int,
) (file.Replacement, error) {
	needPkg := map[string]suggestions.PackageSuggestor{
		"constructor": constructor.Generate,
		"iferr":       iferr.Generate,
	}

	l := loader.New(contents, offset)

	for name, fn := range needPkg {
		t0 := time.Now()
		r, err := fn(l, contents, offset)
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
