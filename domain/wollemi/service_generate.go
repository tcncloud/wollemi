package wollemi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"golang.org/x/mod/modfile"
)

const (
	goLibraryTemplateText = `go_library(
	name = "{{.Name}}",
	srcs = [],
	deps = [],

)
`
	goBinaryTemplateText = `go_binary(
	name = "",
	srcs = [],
	deps = [],
)
	`
	goTestTemplateText = `go_test(
	name = "test",
	srcs = [],
)
	`
	goGetTemplateText = `go_get(
	name = "{{.Name}}",
	revision = "{{.Revision}}",
	get = "{{.Get}}",
	repo = "{{.Repo}}",
	{{if .Install}}
	install = [],
	{{- end}}

)
`
)

var (
	goLibraryTemplate *template.Template
	goBinraryTemplate *template.Template
	goTestTemplate    *template.Template
	goGetTemplate     *template.Template
)

func init() {
	var err error

	goLibraryTemplate = template.New("go_library_template")
	goLibraryTemplate, err = goLibraryTemplate.Parse(goLibraryTemplateText)
	if err != nil {
		log.Fatalf("Can't pase goLibraryTemplate %+v", err)
	}

	goBinraryTemplate = template.New("go_binary_template")
	goBinraryTemplate, err = goBinraryTemplate.Parse(goBinaryTemplateText)
	if err != nil {
		log.Fatalf("Can't pase goBinaryTemplate %+v", err)
	}
	goTestTemplate = template.New("go_test_template")
	goTestTemplate, err = goTestTemplate.Parse(goTestTemplateText)
	if err != nil {
		log.Fatalf("Can't pase goTestTemplate %+v", err)
	}
	goGetTemplate = template.New("go_get_template")
	goGetTemplate, err = goGetTemplate.Parse(goGetTemplateText)
	if err != nil {
		log.Fatalf("Can't pase goGetTemplate %+v", err)
	}
}

type Module struct {
	Path      string       // module path
	Version   string       // module version
	Versions  []string     // available module versions (with -versions)
	Replace   *Module      // replaced by this module
	Time      *time.Time   // time version was created
	Update    *Module      // available update, if any (with -u)
	Main      bool         // is this the main module?
	Indirect  bool         // is this module only an indirect dependency of main module?
	Dir       string       // directory holding files for this module, if any
	GoMod     string       // path to go.mod file used when loading this module, if any
	GoVersion string       // go version used in module
	Error     *ModuleError // error loading module
}

type ModuleError struct {
	Err string // the error itself
}

type PackageError struct {
	ImportStack []string // shortest path from package named on command line to this one
	Pos         string   // position of error (if present, file:line:col)
	Err         string   // the error itself
}
type Package struct {
	Dir           string   // directory containing package sources
	ImportPath    string   // import path of package in dir
	ImportComment string   // path in import comment on package statement
	Name          string   // package name
	Doc           string   // package documentation string
	Target        string   // install path
	Shlib         string   // the shared library that contains this package (only set when -linkshared)
	Goroot        bool     // is this package in the Go root?
	Standard      bool     // is this package part of the standard Go library?
	Stale         bool     // would 'go install' do anything for this package?
	StaleReason   string   // explanation for Stale==true
	Root          string   // Go root or Go path dir containing this package
	ConflictDir   string   // this directory shadows Dir in $GOPATH
	BinaryOnly    bool     // binary-only package (no longer supported)
	ForTest       string   // package is only for use in named test
	Export        string   // file containing export data (when using -export)
	Module        *Module  // info about package's containing module, if any (can be nil)
	Match         []string // command-line patterns matching this package
	DepOnly       bool     // package is only a dependency, not explicitly listed

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
	Incomplete bool            // this package or a dependency has an error
	Error      *PackageError   // error loading package
	DepsErrors []*PackageError // errors loading dependencies

	// please fields
	PleasePath        string
	PleaseDir         string
	IsPackage         bool // true if the entry is part of the project packages
	DependencyInstall []string
	Visibility        []string
}

// GoInfo execute go info command on the given package and return the results in Package format
func GoInfo(pkg string) ([]*Package, error) {
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

	packages := []*Package{}
	json.Unmarshal(js.Bytes(), &packages)

	return packages, nil
}

// GoMod is trying to read the go.mod file from the given location
// and return it parsed into a Mod structure
func GoMod(path string) (*modfile.File, error) {
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

func ProcessPackage(modFile *modfile.File, p *Package, modules map[string]*Package) map[string]*Package {
	if _, ok := modules[p.ImportPath]; !ok {
		// fmt.Print("\033[K") // restore the cursor position and clear the line
		fmt.Printf("\033[K\rQuery package %s", p.ImportPath)
		modules[p.ImportPath] = p
		for _, dep := range p.Imports {
			if strings.Contains(dep, ".") {
				if _, ok = modules[dep]; !ok {
					m, err := GoInfo(dep)
					if err != nil {
						log.Fatalf("Fatal error %+v", err)
					}
					modules = ProcessPackage(modFile, m[0], modules)
				}
			}
		}
		if strings.Contains(p.ImportPath, modFile.Module.Mod.Path) {
			for _, dep := range p.TestImports {
				if strings.Contains(dep, ".") {
					if _, ok = modules[dep]; !ok {
						m, err := GoInfo(dep)
						if err != nil {
							log.Fatalf("Fatal error %+v", err)
						}
						modules = ProcessPackage(modFile, m[0], modules)
					}
				}
			}
			for _, dep := range p.XTestImports {
				if strings.Contains(dep, ".") {
					if _, ok = modules[dep]; !ok {
						m, err := GoInfo(dep)
						if err != nil {
							log.Fatalf("Fatal error %+v", err)
						}
						modules = ProcessPackage(modFile, m[0], modules)
					}
				}
			}
		}
	}
	return modules
}
func ReplaceLast(s, sep, replacement string) string {
	if strings.Contains(s, sep) {
		i := strings.LastIndex(s, sep)
		return s[:i] + replacement + s[i+1:]
	} else {
		return s
	}
}

func (service *Service) ResolveDependencies(pkg string) {

}

// Generate will scan a go project and will generate the build files for a please structure
func (service *Service) Generate(args []string) error {
	os.Setenv("GO111MODULE", "on")

	modFile, err := GoMod("")
	if err != nil {
		service.log.WithError(err).Error("Error reading go.mod file in the local dir")
		return err
	}

	for _, require := range modFile.Require {
		replace := false
		var repl *modfile.Replace
		for _, repl = range modFile.Replace {
			if require.Mod.Path == repl.Old.Path {
				require.Mod.Version = repl.New.Version
				replace = true
				break
			}
		}

		repo := require.Mod.Path
		if replace {
			repo = repl.New.Path
		}

		// if strings.Contains(require.Mod.Version, "v0.0.0-") {
		// use go_get
		// we should try to load the original module definition if exist
		fmt.Printf("go_get(name=\"%s\", revision=\"%s\", repo=\"%s\",)\n", require.Mod.Path, require.Mod.Version, repo)
		// }
		// else {
		// 	// use go_module
		// 	fmt.Printf("go_module(name=\"%s\", revision=\"%s\", repp=\"%s\")\n", require.Mod.Path, require.Mod.Version, repo)
		// }

		// we have to generate
	}

	// info, err := GoInfo("./...")
	// if err != nil {
	// 	return err
	// }
	// modules := map[string]*Package{}
	// for _, m := range info {
	// 	modules = ProcessPackage(modFile, m, modules)
	// }

	// // iterate over modules and compute please path and visibility
	// for k, v := range modules {
	// 	if strings.Contains(k, modFile.Module.Mod.Path) {
	// 		v.IsPackage = true
	// 		v.PleasePath = strings.Replace(k, modFile.Module.Mod.Path, "/", -1)
	// 		if v.PleasePath == "/" {
	// 			v.PleasePath = "//"
	// 		}
	// 		v.PleaseDir = strings.Replace(k, modFile.Module.Mod.Path, ".", -1)
	// 		if v.PleaseDir == "." {
	// 			v.PleaseDir = "./"
	// 		}
	// 	} else if strings.Contains(k, ".") {
	// 		v.IsPackage = false
	// 		//replace last / with :
	// 		v.PleasePath = "//third_party/go/" + ReplaceLast(v.Module.Path, "/", ":")
	// 	}
	// }
	// for k, v := range modules {
	// 	if strings.Contains(k, modFile.Module.Mod.Path) {
	// 		fmt.Printf("\rGenerate build file for package %s %s %s", k, v.PleasePath, v.PleaseDir)
	// 		if len(v.GoFiles) > 0 {
	// 			if v.Name == "main" {
	// 				fmt.Printf(" go_binary package")
	// 				// goBinraryTemplate.Execute(os.Stdout, map[string]interface{}{
	// 				// "Name": v.Name,
	// 				// "Srcs": v.GoFiles,
	// 				// "Deps": v.Imports,
	// 				// })
	// 			} else {
	// 				fmt.Printf(" go_library package")
	// 			}
	// 		}
	// 		if len(v.TestGoFiles) > 0 {
	// 			// test in the same package
	// 			fmt.Printf(" internal test")
	// 		}
	// 		if len(v.XTestGoFiles) > 0 {
	// 			// test in the `package`_test
	// 			fmt.Printf(" external test")
	// 		}
	// 		fmt.Println()
	// 	} else if strings.Contains(k, ".") {
	// 		fmt.Printf("\rGenerate build file for dep %s %s %s\n", k, v.PleasePath, v.PleaseDir)
	// 		// fmt.Printf("\033[K\rGenerate build file for dependency %s", k)
	// 	}
	// }

	return nil

}
