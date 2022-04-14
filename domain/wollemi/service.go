package wollemi

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/tcncloud/wollemi/ports/golang"
	"github.com/tcncloud/wollemi/ports/logging"
	"github.com/tcncloud/wollemi/ports/please"
	"github.com/tcncloud/wollemi/ports/wollemi"
)

type Config = wollemi.Config
type Gofmt = wollemi.Gofmt

func New(
	log logging.Logger,
	filesystem wollemi.Filesystem,
	golang golang.Importer,
	please please.Builder,
	root string,
	wd string,
	gosrc string,
	gopkg string,
) *Service {
	return &Service{
		log:        log,
		filesystem: filesystem,
		golang:     golang,
		please:     please,
		root:       root,
		wd:         wd,
		gopkg:      gopkg,
		gosrc:      gosrc,
	}
}

type Service struct {
	log        logging.Logger
	filesystem wollemi.Filesystem
	golang     golang.Importer
	please     please.Builder
	config     wollemi.Config
	root       string
	wd         string
	gosrc      string
	gopkg      string

	goFormat *goFormat
}

func (this *Service) validateAbsolutePaths(paths []string) error {
	for _, path := range paths {
		if filepath.IsAbs(path) && !strings.HasPrefix(path, this.root) {
			return fmt.Errorf("absolute paths must be under the repo root")
		}
	}
	return nil
}

func (this *Service) normalizePaths(paths []string) []string {
	if len(paths) == 0 {
		paths = []string{"..."}
	}

	for i, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(this.wd, path)
		}
		if strings.HasPrefix(path, this.root) {
			path, _ = filepath.Rel(this.root, path)
		}
		paths[i] = path
	}

	return paths
}

func (this *Service) FindBuildFile(dir string) (string, error) {
	infos, err := this.filesystem.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var paths []string

	for _, info := range infos {
		name := info.Name()
		if isBuildFile(name) {
			paths = append(paths, filepath.Join(dir, name))
		}
	}

	var path string

	switch len(paths) {
	case 0:
		err = os.ErrNotExist
	case 1:
		path = paths[0]
	default:
		err = fmt.Errorf("ambiguous build files: (%s)", strings.Join(paths, ", "))
	}

	return path, err
}

func (this *Service) GoSrcPath(elem ...string) string {
	return filepath.Join(this.gosrc, filepath.Join(elem...))
}

func (this *Service) GoPkgPath(elem ...string) string {
	return filepath.Join(this.gopkg, filepath.Join(elem...))
}

func (this *Service) ReadDir(path string) (*Directory, error) {
	infos, err := this.filesystem.ReadDir(path)
	if err != nil {
		return nil, err
	}

	dir := &Directory{
		Files: make(map[string]os.FileInfo, len(infos)),
		Path:  path,
		Ok:    true,
	}

	for _, info := range infos {
		name := info.Name()

		dir.Files[name] = info

		if info.IsDir() {
			continue
		}

		if isBuildFile(name) {
			dir.BuildFiles = append(dir.BuildFiles, name)
		}

		if filepath.Ext(name) == ".go" {
			dir.GoFiles = append(dir.GoFiles, name)
			dir.HasGoFile = true
		}
	}

	return dir, nil
}

func (this *Service) ReadDirs(out chan *Directory, paths ...string) error {
	in := make(chan string, 1000)

	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for path := range in {
				recursive := filepath.Base(path) == "..."
				if recursive {
					path = filepath.Dir(path)
				}

				dir, err := this.ReadDir(path)
				if err != nil {
					wg.Done()
					continue
				}

				for name, info := range dir.Files {
					if info.IsDir() {
						if name != "." && strings.HasPrefix(name, ".") {
							continue
						}

						if name == "plz-out" {
							continue
						}

						path := filepath.Join(dir.Path, name)
						if recursive {
							path = filepath.Join(path, "...")
						}

						wg.Add(1)
						select {
						case in <- path:
						default:
							go func() {
								in <- path
							}()
						}
					}
				}

				out <- dir
				wg.Done()
			}
		}()
	}

	go func() {
		for _, path := range paths {
			wg.Add(1)
			in <- path
		}

		wg.Wait()
		close(in)
		close(out)
	}()

	return nil
}

func (this *Service) ParseDir(buf *bytes.Buffer, dir *Directory) *Directory {
	log := this.log.WithField("path", dir.Path)

	if len(dir.BuildFiles) > 1 {
		msg := "ambiguous build files: (%s)"
		log.WithError(fmt.Errorf(msg, strings.Join(dir.BuildFiles, ", "))).
			Warn("could not parse dir")

		dir.Ok = false

		return dir
	}

	var buildPath string

	if len(dir.BuildFiles) == 0 {
		buildPath = filepath.Join(dir.Path, BUILD_FILE)
	} else {
		buildPath = filepath.Join(dir.Path, dir.BuildFiles[0])

		if err := this.filesystem.ReadAll(buf, buildPath); err != nil {
			log.WithError(err).Warn("could not read build file")
			dir.Ok = false

			return dir
		}

		build, err := this.please.Parse(buildPath, buf.Bytes())
		if err != nil {
			log.WithError(err).Warn("could not parse build file")
			dir.Ok = false

			return dir
		}

		dir.Build = build
	}

	if dir.HasGoFile && dir.Rewrite && dir.InRunPath {
		gopkg, err := this.golang.ImportDir(dir.Path, dir.GoFiles)
		if err == nil {
			dir.Gopkg = gopkg
		} else if !noBuildableGoSources(err) {
			log.WithError(err).Warn("could not build go import directory")
		}

		if dir.Build == nil {
			dir.Build = this.please.NewFile(buildPath)
		}
	}

	dir.Ok = dir.Build != nil

	return dir
}

type Directory struct {
	Path       string                 `json:"path,omitempty"`
	Rule       string                 `json:"-"`
	Gopkg      *golang.Package        `json:"gopkg,omitempty"`
	Build      please.File            `json:"-"`
	Ok         bool                   `json:"-"`
	Rewrite    bool                   `json:"-"`
	InRunPath  bool                   `json:"-"`
	Files      map[string]os.FileInfo `json:"-"`
	GoFiles    []string               `json:"-"`
	BuildFiles []string               `json:"-"`
	HasGoFile  bool                   `json:"-"`
}

func (Directory) String() string {
	return "{}"
}
