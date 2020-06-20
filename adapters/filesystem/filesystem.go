package filesystem

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tcncloud/wollemi/ports/filesystem"
	"github.com/tcncloud/wollemi/ports/logging"
)

type Config = filesystem.Config

func NewFilesystem(log logging.Logger) *Filesystem {
	return &Filesystem{
		configs: make(map[string]*filesystem.Config),
		log:     log,
	}
}

type Filesystem struct {
	configs map[string]*filesystem.Config
	log     logging.Logger
	mu      sync.Mutex
}

func (*Filesystem) Chdir(dir string) error {
	return os.Chdir(dir)
}

func (*Filesystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (*Filesystem) Remove(path string) error {
	return os.Remove(path)
}

func (*Filesystem) WriteFile(path string, data []byte, mode os.FileMode) error {
	return ioutil.WriteFile(path, data, mode)
}

func (*Filesystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (*Filesystem) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(path)
}

func (*Filesystem) ReadAll(buf *bytes.Buffer, path string) error {
	buf.Reset()

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = buf.ReadFrom(file)

	return err
}

func (*Filesystem) Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

func (this *Filesystem) Config(path string) *filesystem.Config {
	this.mu.Lock()
	defer this.mu.Unlock()

	config, ok := this.configs[path]
	if ok {
		return config
	}

	config = &filesystem.Config{}
	dirs := strings.Split(path, "/")

	var buf bytes.Buffer

	for i := 0; i < len(dirs)+1; i++ {
		path := filepath.Join(dirs[:i]...)

		if tmp, ok := this.configs[path]; ok {
			config = config.Merge(tmp)
			continue
		}

		log := this.log.WithField("path", path).
			WithField("file", ".wollemi.json")

		err := this.ReadAll(&buf, filepath.Join(path, ".wollemi.json"))
		if err == nil {
			tmp := &filesystem.Config{}
			if err := json.Unmarshal(buf.Bytes(), tmp); err != nil {
				log.WithError(err).Warn("could not unmarshal json")

				continue
			}

			config = config.Merge(tmp)

			this.configs[path] = tmp
		}

		if os.IsNotExist(err) {
			this.configs[path] = config
			continue
		}

		if err != nil {
			log.WithError(err).Warn("could not read file")
		}
	}

	return config
}

func (*Filesystem) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

func (*Filesystem) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (*Filesystem) MkdirAll(path string, mode os.FileMode) error {
	return os.MkdirAll(path, mode)
}

func (*Filesystem) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(path)
}
