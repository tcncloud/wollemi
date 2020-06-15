package golang

type Importer interface {
	ImportDir(string) (*Package, error)
}

type Package struct {
	GoFiles      []string
	Goroot       bool
	Imports      []string
	Name         string
	TestGoFiles  []string
	TestImports  []string
	XTestGoFiles []string
	XTestImports []string
}
