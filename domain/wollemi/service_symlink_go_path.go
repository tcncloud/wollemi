package wollemi

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/tcncloud/wollemi/ports/please"
)

func (this *Service) SymlinkGoPath(force bool, paths []string) error {
	if err := this.validateAbsolutePaths(paths); err != nil {
		return err
	}

	paths = this.normalizePaths(paths)

	deps, err := this.please.QueryDeps(paths...)
	if err != nil {
		return err
	}

	symlinks := make(map[string]string)

	for _, dep := range deps {
		if strings.Contains(dep, "#") {
			continue
		}

		if !strings.HasPrefix(dep, "//third_party/go/") {
			continue
		}

		target := please.Split(dep[17:])

		symlinks[target.Path] = dep
	}

	type Symlink struct {
		GoPkg  string
		Target string
	}

	ch := make(chan *Symlink)

	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU()-1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for symlink := range ch {
				gopkg := symlink.GoPkg
				target := symlink.Target

				log := this.log.WithField("go_package", gopkg).
					WithField("build_target", target)

				orig := this.GoSrcPath(gopkg)
				dest := filepath.Join(this.root, "plz-out/gen/third_party/go", gopkg, "src", gopkg)

				err = (func() error {
					err := this.filesystem.Walk(orig, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							if os.IsNotExist(err) {
								return nil
							} else {
								return err
							}
						}

						if !info.IsDir() {
							info, err := this.filesystem.Lstat(path)
							if err != nil {
								return err
							}

							if info.Mode()&os.ModeSymlink != 0 {
								return nil
							}

							if !force {
								path = strings.TrimPrefix(path, this.gosrc)
								log = log.WithField("file", path)
							}

							return fmt.Errorf("file in symlink gopath")
						}

						return nil
					})

					if err != nil {
						if !force {
							return err
						}

						if err := this.filesystem.RemoveAll(orig); err != nil {
							return err
						}
					}

					err = this.filesystem.MkdirAll(orig, os.FileMode(0744))
					if err != nil {
						return err
					}

					var dirs []string
					var files bool

					err = this.filesystem.Walk(dest, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}

						name := info.Name()

						if info.IsDir() {
							if path == dest {
								return nil
							}

							dirs = append(dirs, name)

							return filepath.SkipDir
						}

						if files || filepath.Ext(name) == ".go" {
							files = true
						}

						return nil
					})

					if err != nil {
						return err
					}

					if files {
						if err := this.filesystem.Remove(orig); err != nil {
							return err
						}

						dirs = []string{filepath.Base(orig)}
						orig = filepath.Dir(orig)
						dest = filepath.Dir(dest)
					}

					for _, name := range dirs {
						orig := filepath.Join(orig, name)
						dest := filepath.Join(dest, name)

						link, err := this.filesystem.Readlink(orig)
						if link == dest {
							continue
						}

						if link != "" {
							if err := this.filesystem.Remove(orig); err != nil {
								return err
							}
						}

						if err = this.filesystem.Symlink(dest, orig); err != nil {
							return err
						}
					}

					return nil
				}())

				if err != nil {
					log.WithError(err).Warn("could not symlink")
				} else {
					log.Info("symlinked")
				}
			}
		}()
	}

	for gopkg, target := range symlinks {
		ch <- &Symlink{GoPkg: gopkg, Target: target}
	}

	close(ch)
	wg.Wait()

	return nil
}
