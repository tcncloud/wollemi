package ctl

import (
	"github.com/tcncloud/wollemi/ports/logging"
)

type Application interface {
	Logger() logging.Logger
	Wollemi() (Wollemi, error)
}

type Wollemi interface {
	Format([]string) error
	GoFormat(bool, []string) error
	GoPkgPath(...string) string
	GoSrcPath(...string) string
	SymlinkList(string, bool, bool, []string, []string)
	SymlinkGoPath(bool, []string) error
	RulesUnused(bool, []string, []string, []string) error
}
