package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/cszczepaniak/go-tools/internal"
	"github.com/cszczepaniak/go-tools/internal/file"
)

type foobar struct {
	a    string
	b    int
	C    []byte
	D, E *sync.Mutex
}

type (
	Baz struct {
		a    string
		b    int
		C    []byte
		D, E *sync.Mutex
	}

	Qux struct {
		x, y, z int
	}
)

func main() {
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

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		panic(err)
	}

	col, err := strconv.Atoi(colStr)
	if err != nil {
		panic(err)
	}

	repl, err := internal.GenerateReplacements(filePath, file.Position{
		Line: line,
		Col:  col,
	})
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
