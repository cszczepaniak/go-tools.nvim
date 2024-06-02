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
	"github.com/cszczepaniak/go-tools/internal/logging"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(homeDir, ".go-tools")
	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile(
		filepath.Join(dir, "log.txt"),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0o666,
	)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	logging.InitLogger(logFile)

	fileContents, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	if len(os.Args) < 2 {
		logging.Fatal("must provide one arg")
	}

	parts := strings.Split(os.Args[1], ",")
	if len(parts) != 2 {
		logging.Fatal("argument must be of the form: filename,byte_offset")
	}

	filePath := parts[0]
	byteOffsetStr := parts[1]

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		logging.WithError(err).Fatal("error getting absolute path of file")
	}

	byteOffset, err := strconv.Atoi(byteOffsetStr)
	if err != nil {
		logging.WithError(err).Fatal("error converting line to int")
	}

	repl, err := internal.GenerateReplacements(
		file.Contents{
			AbsPath:  absPath,
			Contents: fileContents,
		},
		byteOffset,
	)
	if err != nil {
		logging.WithError(err).Fatal("error generating replacements")
	}

	if len(repl.Lines) == 0 {
		return
	}

	err = json.NewEncoder(os.Stdout).Encode(repl)
	if err != nil {
		logging.WithError(err).Fatal("error encoding replacement to JSON")
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
