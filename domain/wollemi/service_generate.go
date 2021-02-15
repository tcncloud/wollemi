package wollemi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

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

type Package = build.Package

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
