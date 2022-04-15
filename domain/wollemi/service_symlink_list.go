package wollemi

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tcncloud/wollemi/ports/please"
)

func (this *Service) SymlinkList(name string, broken, prune bool, exclude, include []string) error {
	if err := this.validateAbsolutePaths(include); err != nil {
		return err
	}
	include = this.normalizePaths(include)

	for _, targetPath := range include {
		target := please.Split(targetPath)

		err := this.filesystem.Walk(target.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				if target.Name != "..." && path != target.Path {
					return filepath.SkipDir
				}

				if info.Name() != "." && strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}

				if info.Name() == "plz-out" {
					return filepath.SkipDir
				}

				path := strings.TrimPrefix(path, this.GoSrcPath()+"/")
				for _, prefix := range exclude {
					if strings.HasPrefix(path, prefix) {
						return filepath.SkipDir
					}
				}

				return nil
			}

			ok, err := filepath.Match(name, info.Name())
			if err != nil {
				this.log.WithError(err).
					WithField("name", info.Name()).
					WithField("pattern", name).
					Warn("could not match name")

				return nil
			} else if !ok {
				return nil
			}

			if info.Mode()&os.ModeSymlink != 0 {
				link, err := this.filesystem.Readlink(path)
				if err != nil && !os.IsNotExist(err) {
					this.log.WithError(err).
						WithField("link", strings.TrimPrefix(path, this.GoSrcPath()+"/")).
						Warn("could not read link")

					return nil
				}

				log := this.log.
					WithField("link", strings.TrimPrefix(path, this.GoSrcPath()+"/")).
					WithField("path", strings.TrimPrefix(link, this.root+"/"))

				if broken {
					_, err := this.filesystem.Stat(link)
					if err == nil {
						return nil
					}

					if !os.IsNotExist(err) {
						log.WithError(err).Warn("could not stat link")
						return nil
					}
				}

				if prune {
					if err := this.filesystem.Remove(path); err != nil {
						log.WithError(err).Warn("could not remove link")
						return nil
					}

					log.Info("symlink deleted")
				} else {
					log.Info("symlink")
				}
			}

			return nil
		})

		if err != nil {
			this.log.WithError(err).
				WithField("path", target.Path).
				Warn("could not walk")
		}
	}
	return nil
}
