package wollemi

import (
	"path/filepath"
	"strings"
)

// split splits a target into path and name components.
func split(target string) (string, string) {
	target = strings.TrimPrefix(target, "//")

	colon := strings.LastIndex(target, ":")
	if colon < 0 {
		base := filepath.Base(target)
		switch base {
		case "...":
			return filepath.Dir(target), base
		default:
			return target, base
		}
	}

	return target[:colon], target[colon+1:]
}

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

func inRunPath(target string, run ...string) bool {
	for _, path := range run {
		if path == "..." {
			return true
		}

		targetPath, _ := split(target)

		if filepath.Base(path) != "..." {
			if path == targetPath {
				return true
			} else {
				continue
			}
		}

		dir := filepath.Dir(path)

		if strings.HasPrefix(targetPath, dir) {
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
