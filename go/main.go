package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cszczepaniak/go-tools/internal"
	"github.com/cszczepaniak/go-tools/internal/file"
)

func main() {
	fileContents, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		panic("must provide one arg")
	}

	parts := strings.Split(os.Args[1], ",")
	if len(parts) != 3 {
		panic("arg must be of the form: filename,linenumber,colnumber")
	}

	filePath := parts[0]
	lineStr := parts[1]
	colStr := parts[2]

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		panic(err)
	}

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		panic(err)
	}

	col, err := strconv.Atoi(colStr)
	if err != nil {
		panic(err)
	}

	repl, err := internal.GenerateReplacements(
		file.Contents{
			AbsPath:  absPath,
			Contents: fileContents,
		},
		file.Position{
			Line: line,
			Col:  col,
		},
	)
	if err != nil {
		panic(err)
	}

	if len(repl.Lines) == 0 {
		return
	}

	err = json.NewEncoder(os.Stdout).Encode(repl)
	if err != nil {
		panic(err)
	}
}

func fooBar() (int, file.Range, error) {
	_, err := iAmFallible()
	if err != nil {
		_, err = iAmFallible()
		if err != nil {
			return 0, file.Range{}, err
		}
		return 0, file.Range{}, err
	}

	foo := func() (bool, int, error) {
		_, err := iAmFallible()
		_ = err
		return false, 0, nil
	}
	_ = foo

	return 0, file.Range{}, nil
}

func iAmFallible() (int, error) {
	return 0, nil
}

func iAmFallible2() (int, error) {
	return 0, nil
}
