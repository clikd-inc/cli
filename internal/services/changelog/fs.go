package changelog

import (
	"io"
	"os"
)

// FileSystem ...
type FileSystem interface {
	Exists(path string) bool
	MkdirP(path string) error
	Create(name string) (File, error)
	WriteFile(path string, content []byte) error
}

// File ...
type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Writer
	Stat() (os.FileInfo, error)
}

// FS provides file system operations
var FS FileSystem = &osFileSystem{}

// SetFS ermöglicht es, das Dateisystem für Tests zu ersetzen
func SetFS(fs FileSystem) {
	FS = fs
}

type osFileSystem struct{}

func (*osFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (*osFileSystem) MkdirP(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		//nolint:gosec
		return os.MkdirAll(path, os.ModePerm)
	}
	return nil
}

func (*osFileSystem) Create(name string) (File, error) {
	//nolint: gosec
	return os.Create(name)
}

func (*osFileSystem) WriteFile(path string, content []byte) error {
	//nolint:gosec
	return os.WriteFile(path, content, os.ModePerm)
}
