package comm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func DefaultFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".go-tools", "state", "daemon.txt"), nil
}

type fileCommunication struct {
	f *os.File
	r *bufio.Reader
	w *fsnotify.Watcher
}

func (fc fileCommunication) Close() error {
	return fc.f.Close()
}

func NewFileCommunication(name string) (fileCommunication, error) {
	dir, _ := filepath.Split(name)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return fileCommunication{}, err
	}

	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o666)
	if err != nil {
		return fileCommunication{}, err
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fileCommunication{}, err
	}

	err = w.Add(f.Name())
	if err != nil {
		return fileCommunication{}, err
	}

	return fileCommunication{
		f: f,
		r: bufio.NewReader(f),
		w: w,
	}, nil
}

func (fc fileCommunication) ReadLine() (string, bool) {
	for {
		ln, err := fc.r.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fc.waitForWrite()
				continue
			}

			return "", false
		}

		return ln[:len(ln)-1], true
	}
}

func (fc fileCommunication) WriteLine(s string) error {
	_, err := fmt.Fprintln(fc.f, s)
	return err
}

func (fc fileCommunication) waitForWrite() {
	for ev := range fc.w.Events {
		if ev.Name == fc.f.Name() && ev.Has(fsnotify.Write) {
			return
		}
	}
}
