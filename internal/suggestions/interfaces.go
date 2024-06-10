package suggestions

import (
	"github.com/cszczepaniak/go-tools/internal/file"
	"github.com/cszczepaniak/go-tools/internal/loader"
	"golang.org/x/tools/go/packages"
)

type FileParser interface {
	ParseFile() (loader.File, error)
}

type PackageLoader interface {
	FileParser
	LoadPackage() (*packages.Package, error)
}

type FileSuggestor func(FileParser, file.Contents, int) (file.Replacement, error)
type PackageSuggestor func(PackageLoader, file.Contents, int) (file.Replacement, error)
