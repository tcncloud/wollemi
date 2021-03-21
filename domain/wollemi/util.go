package wollemi

import (
	"path/filepath"
	"strings"

	"github.com/tcncloud/wollemi/ports/please"
)

// noBuildableGoSources determines if this error occurred because no
// buildable golang sources were detected in the package.
func noBuildableGoSources(err error) bool {
	switch err {
	case nil:
		return false
	default:
		return strings.Contains(err.Error(), "no buildable Go source files")
	}
}

func inRunPath(targetPath string, run ...string) bool {
	for _, path := range run {
		if path == "..." {
			return true
		}

		target := please.Split(targetPath)

		if filepath.Base(path) != "..." {
			if path == target.Path {
				return true
			} else {
				continue
			}
		}

		if strings.HasPrefix(target.Path, filepath.Dir(path)) {
			return true
		}
	}

	return false
}

// nonBlockingSend ensures a non blocking send of directory to channel. It will
// first attempt to send as normal on a select but if that blocks as a last
// resort it will send inside of a goroutine. This is not expected to happen
// but is a failsafe to prevent potential deadlocks.
func nonBlockingSend(ch chan *Directory, dir *Directory) {
	select {
	case ch <- dir:
	default:
		go func() { ch <- dir }()
	}
}

func isBuildFile(name string) bool {
	for _, want := range []string{"BUILD.plz", "BUILD"} {
		if name == want {
			return true
		}
	}

	return false
}
