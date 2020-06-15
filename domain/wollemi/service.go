package wollemi

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/tcncloud/wollemi/ports/golang"
	"github.com/tcncloud/wollemi/ports/logging"
	"github.com/tcncloud/wollemi/ports/please"
	"github.com/tcncloud/wollemi/ports/wollemi"
)

func New(
	log logging.Logger,
	filesystem wollemi.Filesystem,
	golang golang.Importer,
	please please.Builder,
	gosrc string,
	gopkg string,
) *Service {
	return &Service{
		log:        log,
		filesystem: filesystem,
		golang:     golang,
		please:     please,
		gopkg:      gopkg,
		gosrc:      gosrc,
	}
}

type Service struct {
	log        logging.Logger
	filesystem wollemi.Filesystem
	golang     golang.Importer
	please     please.Builder
	gosrc      string
	gopkg      string
}

func (this *Service) GoPkgPath(elem ...string) string {
	return this.GoSrcPath(this.gopkg, filepath.Join(elem...))
}

func (this *Service) GoSrcPath(elem ...string) string {
	return filepath.Join(this.gosrc, filepath.Join(elem...))
}

func (this *Service) Walk(out chan *Directory, paths ...string) error {
	defer close(out)

	for _, path := range paths {
		if filepath.Base(path) != "..." {
			out <- &Directory{Path: path, Ok: true}
			continue
		}

		dir := filepath.Dir(path)
		err := this.filesystem.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			name := info.Name()

			if info.IsDir() {
				if name != "." && strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}

				if name == "plz-out" {
					return filepath.SkipDir
				}

				out <- &Directory{Path: path, Ok: true}
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Service) Parse(buf *bytes.Buffer, dir *Directory) *Directory {
	log := this.log.WithField("path", dir.Path)

	var parsedPlzBuild bool

	defer func() {
		log.WithField("go_package", dir.Gopkg != nil).
			WithField("build_file", parsedPlzBuild).
			Debug("parsed")

	}()

	if dir.InRunPath && dir.Rewrite {
		gopkg, err := this.golang.ImportDir(dir.Path)
		if err == nil {
			dir.Gopkg = gopkg
		} else if !noBuildableGoSources(err) {
			log.WithError(err).Warn("could not build go import directory")
		}
	}

	buildPath := filepath.Join(dir.Path, "BUILD.plz")

	if err := this.filesystem.ReadAll(buf, buildPath); err != nil {
		if !os.IsNotExist(err) {
			dir.Ok = false
			log.WithError(err).Warn("could not read build file")
		} else if dir.Gopkg != nil {
			generate := len(dir.Gopkg.GoFiles) > 0 ||
				len(dir.Gopkg.TestGoFiles) > 0 ||
				len(dir.Gopkg.XTestGoFiles) > 0

			if generate {
				dir.Build = this.please.NewFile(buildPath)
			} else {
				dir.Ok = false
			}
		} else {
			dir.Ok = false
		}

		return dir
	}

	file, err := this.please.Parse(buildPath, buf.Bytes())
	if err != nil {
		log.WithError(err).Warn("could not parse build file")
		dir.Ok = false

		return dir
	} else {
		parsedPlzBuild = true
	}

	dir.Build = file

	return dir
}

type Directory struct {
	Path      string
	Rule      string
	Gopkg     *golang.Package
	Build     please.File
	Ok        bool
	Rewrite   bool
	InRunPath bool
}
