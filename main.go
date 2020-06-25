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

	gosrc := filepath.Join(golang.GOPATH(), "src") + "/"

	var gopkg string

	buf := bytes.NewBuffer(nil)
	if err := filesystem.ReadAll(buf, filepath.Join(root, "go.mod")); err == nil {
		gopkg = golang.ModulePath(buf.Bytes())
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("could not read go.mod: %v", err)
	}

	if gopkg == "" {
		if !strings.HasPrefix(root, gosrc) {
			return nil, fmt.Errorf("project root must have go.mod or exist in go path")
		}

		gopkg = strings.TrimPrefix(root, gosrc)
	}

	bazel := bazel.NewBuilder(log, please.NewCtl(), filesystem)

	log.WithField("working_directory", wd).
		WithField("project_root", root).
		WithField("go_package", gopkg).
		WithField("go_path", golang.GOPATH()).
		WithField("go_root", golang.GOROOT()).
		Debug("wollemi initialized")

	return wollemi.New(log, filesystem, golang, bazel, root, wd, gosrc, gopkg), nil
}
