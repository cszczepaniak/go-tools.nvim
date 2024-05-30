package internal

import (
	"bufio"
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOmni(t *testing.T) {
	entries, err := os.ReadDir("./testdata")
	require.NoError(t, err)

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") {
			continue
		}

		fullPath := path.Join("./testdata", e.Name())
		exps := getExpectations(t, fullPath)

		absPath, err := filepath.Abs(fullPath)
		require.NoError(t, err)

		fileBytes, err := os.ReadFile(fullPath)
		require.NoError(t, err)

		contents := file.Contents{
			AbsPath:  absPath,
			Contents: fileBytes,
		}

		t.Run(e.Name(), func(t *testing.T) {
			for _, exp := range exps {
				t.Run(exp.name, func(t *testing.T) {
					for _, p := range exp.positions {
						r, err := GenerateReplacements(contents, p)
						require.NoError(t, err)
						assert.Equal(t, exp.exp.Range, r.Range)
						assert.Equal(
							t,
							strings.Join(exp.exp.Lines, "\n"),
							strings.Join(r.Lines, "\n"),
						)
					}
				})
			}
		})
	}
}

type testExpectation struct {
	name      string
	positions []file.Position
	exp       file.Replacement
}

func getExpectations(t *testing.T, filepath string) []testExpectation {
	t.Helper()

	const (
		startPosTag = `// START`
		endPosTag   = ` // END`
		startExpTag = `/* EXPECT`
		endExpTag   = `*/`
	)

	bs, err := os.ReadFile(filepath)
	require.NoError(t, err)

	var exps []testExpectation
	updateCurrent := func(fn func(exp testExpectation) testExpectation) {
		updated := fn(exps[len(exps)-1])
		exps[len(exps)-1] = updated
	}

	parsingInput := false
	parsingExp := false
	ln := 0
	sc := bufio.NewScanner(bytes.NewReader(bs))
	for sc.Scan() {
		ln++

		if idx := strings.Index(sc.Text(), startPosTag); !parsingExp && idx != -1 {
			exps = append(exps, testExpectation{
				name: strings.TrimPrefix(sc.Text(), startPosTag+" "),
				positions: []file.Position{{
					Line: ln + 1,
					Col:  idx + 1, // columns are 1-based
				}},
				exp: file.Replacement{
					Range: file.Range{
						Start: file.Position{
							Line: ln + 1,
							Col:  idx + 1,
						},
					},
				},
			})
			parsingInput = true
			continue
		}

		if idx := strings.Index(sc.Text(), endPosTag); !parsingExp && parsingInput && idx != -1 {
			updateCurrent(func(exp testExpectation) testExpectation {
				exp.positions = append(exp.positions, file.Position{
					Line: ln,
					Col:  idx + 1,
				})
				exp.exp.Range.Stop = file.Position{
					Line: ln,
					Col:  idx + 1, // columns are 1-based
				}
				return exp
			})
			parsingInput = false
			continue
		}

		if sc.Text() == startExpTag {
			parsingExp = true
			continue
		}

		if sc.Text() == endExpTag {
			parsingExp = false
			continue
		}

		if parsingExp {
			updateCurrent(func(exp testExpectation) testExpectation {
				exp.exp.Lines = append(exp.exp.Lines, sc.Text())
				return exp
			})
		}
	}

	return exps
}
