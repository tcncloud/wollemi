package wollemi

import (
	"bytes"
	"os"
	"path/filepath"
)

type Filesystem interface {
	Stat(string) (os.FileInfo, error)
	Lstat(string) (os.FileInfo, error)
	Config(string) Config
	Walk(string, filepath.WalkFunc) error
	ReadAll(*bytes.Buffer, string) error
	ReadDir(string) ([]os.FileInfo, error)
	Readlink(string) (string, error)
	Symlink(string, string) error
	RemoveAll(string) error
	Remove(string) error
	MkdirAll(string, os.FileMode) error
}
