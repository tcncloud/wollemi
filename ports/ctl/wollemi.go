package ctl

import (
	"github.com/tcncloud/wollemi/ports/logging"
	"github.com/tcncloud/wollemi/ports/wollemi"
)

type Application interface {
	Logger() logging.Logger
	Wollemi() (Wollemi, error)
}

type Wollemi interface {
	Format(wollemi.Config, []string) error
	GoFormat(wollemi.Config, []string) error
	GoPkgPath(...string) string
	GoSrcPath(...string) string
	SymlinkList(string, bool, bool, []string, []string) error
	SymlinkGoPath(bool, []string) error
	RulesUnused(bool, []string, []string, []string) error
}
