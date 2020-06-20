package golang

type Importer interface {
	ImportDir(string, []string) (*Package, error)
	IsGoroot(string) bool
	GOPATH() string
}

type Package struct {
	GoFiles        []string `json:"go_files,omitempty"`
	Goroot         bool     `json:"goroot,omitempty"`
	Imports        []string `json:"imports,omitempty"`
	Name           string   `json:"name,omitempty"`
	TestGoFiles    []string `json:"test_go_files,omitempty"`
	TestImports    []string `json:"test_imports,omitempty"`
	XTestGoFiles   []string `json:"x_test_go_files,omitempty"`
	XTestImports   []string `json:"x_test_imports,omitempty"`
	IgnoredGoFiles []string `json:"ignored_go_files,omitempty"`
}
