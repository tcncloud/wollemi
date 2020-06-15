package please

import (
	"os"
)

type Filesystem interface {
	Remove(string) error
	ReadDir(string) ([]os.FileInfo, error)
	WriteFile(string, []byte, os.FileMode) error
}
