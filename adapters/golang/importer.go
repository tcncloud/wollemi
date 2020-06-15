package golang

import (
	"go/build"
	"os"
	"sync"

	"github.com/tcncloud/wollemi/ports/golang"
)

type Package = golang.Package

func init() {
	os.Setenv("GO111MODULE", "OFF")
}

func NewImporter() *Importer {
	return &Importer{
		pkgs: make(map[string]*Package),
	}
}

type Importer struct {
	pkgs map[string]*Package
	mu   sync.Mutex
	wd   string
}

func (this *Importer) ImportDir(dir string) (*Package, error) {
	pkg, err := NewPackage(build.ImportDir(dir, 0))
	if err != nil {
		return nil, err
	}

	this.mu.Lock()
	this.pkgs[dir] = pkg // TODO: make dir package path? (extra work done otherwise?)
	pkg.Imports = this.removeGorootImports(pkg.Imports)
	pkg.TestImports = this.removeGorootImports(pkg.TestImports)
	pkg.XTestImports = this.removeGorootImports(pkg.XTestImports)
	this.mu.Unlock()

	return pkg, nil
}

func (this *Importer) removeGorootImports(imports []string) []string {
	result := make([]string, 0, len(imports))

	var err error

	for _, dep := range imports {
		pkg, ok := this.pkgs[dep]
		if !ok {
			pkg, err = NewPackage(build.Import(dep, "", build.FindOnly))
			if err != nil {
				// Assume non goroot package on error. An error will occur when
				// the package could not be found in the gopath.

				pkg = &Package{Goroot: false}
			}

			this.pkgs[dep] = pkg
		}

		if !pkg.Goroot {
			result = append(result, dep)
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func NewPackage(in *build.Package, err error) (*Package, error) {
	if err != nil {
		return nil, err
	}

	var out *Package

	if in != nil {
		out = &Package{
			Name:         in.Name,
			Goroot:       in.Goroot,
			Imports:      in.Imports,
			TestImports:  in.TestImports,
			XTestImports: in.XTestImports,
			GoFiles:      in.GoFiles,
			TestGoFiles:  in.TestGoFiles,
			XTestGoFiles: in.XTestGoFiles,
		}
	}

	return out, nil
}
