package main

import (
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

	gosrc := filepath.Join(os.Getenv("GOPATH"), "src") + "/"

	if !strings.HasPrefix(wd, gosrc) {
		return nil, fmt.Errorf("working directory not in gopath")
	}

	gopkg := strings.TrimPrefix(wd, gosrc)

	var (
		log        = app.Logger()
		filesystem = filesystem.NewFilesystem(log)
		bazel      = bazel.NewBuilder(log, please.NewCtl(), filesystem)
		golang     = golang.NewImporter()
	)

	return wollemi.New(log, filesystem, golang, bazel, gosrc, gopkg), nil
}
