package please

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type Target struct {
	Path string
	Name string
}

// Less returns true when this target is less than the one provided.
func (t *Target) Less(in *Target) bool {
	return (t.Path == in.Path && t.Name < in.Name) || t.Path < in.Path
}

// String stringifies this target.
func (t *Target) String() string {
	if filepath.Base(t.Path) == t.Name {
		return fmt.Sprintf("//%s", t.Path)
	}

	return fmt.Sprintf("//%s:%s", t.Path, t.Name)
}

// Rel returns the relative path to this target from the provided path.
func (t *Target) Rel(path string) string {
	if t.Path == path {
		return fmt.Sprintf(":%s", t.Name)
	}

	return t.String()
}

// Split splits a please target into path and name components.
func Split(path string) *Target {
	path = strings.TrimPrefix(path, "//")

	colon := strings.LastIndex(path, ":")
	if colon < 0 {
		name := filepath.Base(path)
		switch name {
		case "...":
			return &Target{Path: filepath.Dir(path), Name: name}
		default:
			return &Target{Path: path, Name: name}
		}
	}

	return &Target{Path: path[:colon], Name: path[colon+1:]}
}

// SortDeps sorts please build rule deps.
func SortDeps(deps []string) {
	sort.Slice(deps, func(i, j int) bool {
		return Split(deps[i]).Less(Split(deps[j]))
	})
}
