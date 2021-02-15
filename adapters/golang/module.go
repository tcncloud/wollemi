package golang

import (
	"bytes"
	"encoding/json"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"golang.org/x/mod/modfile"
)

type PackageInfo struct {
	Dir           string      // directory containing package sources
	ImportPath    string      // import path of package in dir
	ImportComment string      // path in import comment on package statement
	Name          string      // package name
	Doc           string      // package documentation string
	Target        string      // install path
	Shlib         string      // the shared library that contains this package (only set when -linkshared)
	Goroot        bool        // is this package in the Go root?
	Standard      bool        // is this package part of the standard Go library?
	Stale         bool        // would 'go install' do anything for this package?
	StaleReason   string      // explanation for Stale==true
	Root          string      // Go root or Go path dir containing this package
	ConflictDir   string      // this directory shadows Dir in $GOPATH
	BinaryOnly    bool        // binary-only package (no longer supported)
	ForTest       string      // package is only for use in named test
	Export        string      // file containing export data (when using -export)
	Module        *ModuleInfo // info about package's containing module, if any (can be nil)
	Match         []string    // command-line patterns matching this package
	DepOnly       bool        // package is only a dependency, not explicitly listed

	// Source files
	GoFiles         []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles        []string // .go source files that import "C"
	CompiledGoFiles []string // .go files presented to compiler (when using -compiled)
	IgnoredGoFiles  []string // .go source files ignored due to build constraints
	CFiles          []string // .c source files
	CXXFiles        []string // .cc, .cxx and .cpp source files
	MFiles          []string // .m source files
	HFiles          []string // .h, .hh, .hpp and .hxx source files
	FFiles          []string // .f, .F, .for and .f90 Fortran source files
	SFiles          []string // .s source files
	SwigFiles       []string // .swig files
	SwigCXXFiles    []string // .swigcxx files
	SysoFiles       []string // .syso object files to add to archive
	TestGoFiles     []string // _test.go files in package
	XTestGoFiles    []string // _test.go files outside package

	// Cgo directives
	CgoCFLAGS    []string // cgo: flags for C compiler
	CgoCPPFLAGS  []string // cgo: flags for C preprocessor
	CgoCXXFLAGS  []string // cgo: flags for C++ compiler
	CgoFFLAGS    []string // cgo: flags for Fortran compiler
	CgoLDFLAGS   []string // cgo: flags for linker
	CgoPkgConfig []string // cgo: pkg-config names

	// Dependency information
	Imports      []string          // import paths used by this package
	ImportMap    map[string]string // map from source import to ImportPath (identity entries omitted)
	Deps         []string          // all (recursively) imported dependencies
	TestImports  []string          // imports from TestGoFiles
	XTestImports []string          // imports from XTestGoFiles

	// Error information
	Incomplete bool                // this package or a dependency has an error
	Error      *PackageInfoError   // error loading package
	DepsErrors []*PackageInfoError // errors loading dependencies
}

type PackageInfoError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error (if present, file:line:col)
	Err         string   // the error itself
}
type ModuleInfo struct {
	Path      string           // module path
	Version   string           // module version
	Versions  []string         // available module versions (with -versions)
	Replace   *ModuleInfo      // replaced by this module
	Time      *time.Time       // time version was created
	Update    *ModuleInfo      // available update, if any (with -u)
	Main      bool             // is this the main module?
	Indirect  bool             // is this module only an indirect dependency of main module?
	Dir       string           // directory holding files for this module, if any
	GoMod     string           // path to go.mod file used when loading this module, if any
	GoVersion string           // go version used in module
	Error     *ModuleInfoError // error loading module
}

type ModuleInfoError struct {
	Err string // the error itself
}

// GoInfo execute go info command on the given package and return the results in Package format
func GoInfo(pkg string) ([]*PackageInfo, error) {
	out, err := exec.Command("go", "list", "-json", pkg).CombinedOutput()
	if err != nil {
		return nil, err
	}

	js := bytes.Buffer{}
	js.WriteByte('[')
	open_bracket := 0
	for _, j := range out {
		js.WriteByte(j)
		if j == '{' {
			open_bracket++
		}
		if j == '}' {
			open_bracket--
			if open_bracket == 0 {
				js.WriteByte(',')
			}
		}
	}
	js.Truncate(js.Len() - 2)
	js.WriteByte(']')

	packages := []*build.Package{}
	json.Unmarshal(js.Bytes(), &packages)

	return packages, nil
}

type GoMod struct {
	fileiname string
}

func NewGoMod(path string) (*GoMod, error) {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}

	ret := &GoMod{}

}

// readGoMod is trying to read the go.mod file from the given location
// and return it parsed into a Mod structure
func readGoMod(path string) (*modfile.File, error) {
	b, err := ioutil.ReadFile(path + "go.mod")
	if err != nil {
		return nil, err
	}
	mod, err := modfile.Parse("go.mod", b, nil)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func isInRequire()

func ProcessGoMod(mod *modfile.File) {
	for _, r := range mod.Require {

	}
}
