package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tcncloud/wollemi/adapters/bazel"
	"github.com/tcncloud/wollemi/adapters/cobra"
	"github.com/tcncloud/wollemi/adapters/filesystem"
	"github.com/tcncloud/wollemi/adapters/golang"
	"github.com/tcncloud/wollemi/adapters/logrus"
	"github.com/tcncloud/wollemi/adapters/please"
	"github.com/tcncloud/wollemi/domain/wollemi"
	"github.com/tcncloud/wollemi/ports/ctl"
	"github.com/tcncloud/wollemi/ports/logging"
)

func main() {
	app := NewApplication()

	if err := cobra.Ctl(app).Execute(); err != nil {
		os.Exit(1)
	}
}

func NewApplication() *Application {
	return &Application{
		logger: logrus.NewLogger(os.Stderr),
	}
}

type Application struct {
	logger logging.Logger
}

func (app *Application) Logger() logging.Logger {
	return app.logger
}

func (app *Application) Wollemi() (ctl.Wollemi, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get working directory: %v", err)
	}

	if path, err := os.Readlink(wd); err == nil {
		wd = path
	}

	log := app.Logger()
	golang := golang.NewImporter()
	filesystem := filesystem.NewFilesystem(log)

	root, ok := (func(path string) (string, bool) {
		for ; path != "/"; path = filepath.Dir(path) {
			_, err := filesystem.Stat(filepath.Join(path, ".plzconfig"))
			if err == nil {
				return path, true
			}
		}

		return "", false
	}(wd))

	if !ok {
		return nil, fmt.Errorf("could not find root .plzconfig")
	}

	if err := filesystem.Chdir(root); err != nil {
		return nil, fmt.Errorf("could not chdir: %v", err)
	}

	var gosrc, gopkg string

	buf := bytes.NewBuffer(nil)
	if err := filesystem.ReadAll(buf, filepath.Join(root, "go.mod")); err == nil {
		gopkg = golang.ModulePath(buf.Bytes())
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("could not read go.mod: %v", err)
	}

	if gopkg == "" {
		for _, gopath := range strings.Split(golang.GOPATH(), ":") {
			gosrc = filepath.Join(gopath, "src") + "/"

			if strings.HasPrefix(root, gosrc) {
				gopkg = strings.TrimPrefix(root, gosrc)
				break
			}
		}
	}

	if gopkg == "" && gosrc != root {
		gosrc = root
	}

	pleaseCtl := please.NewCtl()

	config, err := pleaseCtl.Config(filepath.Join(root, ".plzconfig"))
	if err == nil && config.Go.ImportPath != "" {
		gopkg = config.Go.ImportPath
	}

	bazel := bazel.NewBuilder(log, pleaseCtl, filesystem)

	log.WithField("working_directory", wd).
		WithField("project_root", root).
		WithField("go_package", gopkg).
		WithField("go_path", golang.GOPATH()).
		WithField("go_root", golang.GOROOT()).
		Debug("wollemi initialized")

	return wollemi.New(log, filesystem, golang, bazel, root, wd, gosrc, gopkg), nil
}
