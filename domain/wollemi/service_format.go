package wollemi

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/tcncloud/wollemi/ports/filesystem"
	"github.com/tcncloud/wollemi/ports/please"
)

const (
	BUILD_FILE = "BUILD.plz"
)

func (this *Service) Format(paths []string) error {
	return this.GoFormat(false, paths)
}

func (this *Service) GoFormat(rewrite bool, paths []string) error {
	paths = this.normalizePaths(paths)

	log := this.log.WithField("rewrite", rewrite)
	for i, path := range paths {
		log = log.WithField(fmt.Sprintf("path[%d]", i), path)
	}

	log.Debug("running gofmt")

	collect := make(chan *Directory, 1000)
	parse := make(chan *Directory, 1000)
	walk := make(chan *Directory, 1000)

	directories := make(map[string]*Directory)
	external := make(map[string][]string)
	internal := make(map[string]string)
	genfiles := make(map[string]string)
	isGoroot := make(map[string]bool)

	for i := 0; i < runtime.NumCPU()-1; i++ {
		go func() {
			var buf bytes.Buffer

			for dir := range parse {
				dir.InRunPath = inRunPath(dir.Path, paths...)
				dir.Rewrite = rewrite

				nonBlockingSend(collect, this.ParseDir(&buf, dir))
			}
		}()
	}

	if err := this.ReadDirs(walk, paths...); err != nil {
		return fmt.Errorf("could not walk: %v", err)
	}

	delegated := make(map[string]struct{})
	parsing := 0

	for walk != nil || parsing > 0 {
		select {
		case dir, ok := <-walk:
			if !ok {
				walk = nil
			} else {
				parse <- dir
				parsing++
			}
		case dir := <-collect:
			parsing--

			if !dir.Ok {
				continue
			}

			directories[dir.Path] = dir

			if dir.Gopkg != nil {
				for _, imports := range [][]string{
					dir.Gopkg.Imports,
					dir.Gopkg.TestImports,
					dir.Gopkg.XTestImports,
				} {
				Imports:
					for _, godep := range imports {
						var prefix bool

						goroot, ok := isGoroot[godep]
						if !ok {
							if prefix = strings.HasPrefix(godep, this.gopkg); prefix {
								goroot = false
							} else {
								goroot = this.golang.IsGoroot(godep)
							}

							isGoroot[godep] = goroot
						}

						if goroot {
							continue
						}

						path := godep

						if prefix || strings.HasPrefix(path, this.gopkg) {
							path = strings.TrimPrefix(path, this.gopkg+"/")
						} else {
							path = filepath.Join("third_party/go", path)
						}

						if _, ok := external[godep]; ok {
							continue
						}

						if inRunPath(path, paths...) {
							continue
						}

						chunks := strings.Split(path, "/")

						for i := len(chunks); i > 0; i-- {
							path := filepath.Join(chunks[0:i]...)

							if _, ok := delegated[path]; ok {
								continue Imports
							}

							if _, ok := directories[path]; ok {
								continue Imports
							}

							dir, err := this.ReadDir(path)
							if os.IsNotExist(err) {
								continue
							}

							if err != nil {
								this.log.WithError(err).
									WithField("path", path).
									Warn("could not read dir")

								continue
							}

							if len(dir.BuildFiles) == 0 {
								continue
							}

							delegated[path] = struct{}{}

							parsing++
							parse <- dir
						}
					}
				}
			}

			dir.Build.GetRules(func(rule please.Rule) {
				switch rule.Kind() {
				case "go_copy", "go_mock", "go_library", "go_test", "grpc_library", "proto_library":
					name := rule.AttrString("name")

					target, path := dir.Path, dir.Path
					if filepath.Base(path) != name {
						path = filepath.Join(path, name)
						target += ":" + name
					}

					importPath := rule.AttrString("import_path")
					if importPath != "" {
						external[importPath] = append(external[path], "//"+target)
					} else {
						internal[path] = "//" + target

						if rule.Kind() == "go_copy" {
							genfiles[path+".cp.go"] = target
						}
					}
				case "go_get", "go_get_with_sources":
					get := strings.TrimSuffix(rule.AttrString("get"), "/...")
					if get == "" {
						if install := rule.AttrStrings("install"); len(install) > 0 {
							sort.Strings(install)
							get = strings.TrimSuffix(install[0], "/...")
						}
					}

					name := rule.AttrString("name")

					target := dir.Path
					if filepath.Base(dir.Path) != name {
						target += ":" + name
					}

					if rule.Kind() == "go_get_with_sources" {
						get = rule.AttrStrings("outs")[0]
					}

					if get != "" && rule.AttrLiteral("binary") != "True" {
						external[get] = append(external[get], "//"+target)
					}
				}
			})
		}
	}

	close(parse)

	get := NewChanFunc(1, 0)
	gen := NewChanFunc(runtime.NumCPU()-1, 0)

	getTarget := (func() func(*filesystem.Config, string, bool) (string, string) {
		var inner func(*filesystem.Config, string, bool) (string, string)

		inner = func(config *filesystem.Config, path string, isFile bool) (string, string) {
			if target, ok := config.KnownDependency[path]; ok {
				return target, path
			}

			if isFile && filepath.Ext(path) == ".go" {
				if target, ok := genfiles[path]; ok {
					return target, path
				} else {
					return "", path
				}
			}

			if strings.HasPrefix(path, this.gopkg+"/") {
				relpath := strings.TrimPrefix(path, this.gopkg+"/")

				if target, ok := internal[relpath]; ok {
					return target, path
				}

				return fmt.Sprintf("//%s", relpath), path
			}

			targets, ok := external[path]
			if ok {
				if len(targets) > 1 {
					this.log.WithField("choices", targets).
						WithField("godep", path).
						WithField("chose", targets[0]).
						Warn("ambiguous godep")
				}

				return targets[0], path
			}

			path = filepath.Dir(path)
			if path == "." {
				return "", path
			}

			return inner(config, path, isFile)
		}

		return func(config *filesystem.Config, path string, isFile bool) (string, string) {
			var target string

			get.RunBlock(func() {
				target, path = inner(config, path, isFile)
			})

			return target, path
		}
	}())

	for path, dir := range directories {
		if !dir.InRunPath {
			continue
		}

		log := this.log.WithField("path", path)

		path, dir := path, dir

		var deps []string
		var unresolved []string
		var include []string
		var exclude []string
		var targets []string
		var goFiles []string
		var imports []string

		gen.Run(func() {
			config := this.filesystem.Config(path)

			rulesByKind := make(map[string][]please.Rule)
			isGeneratedRule := make(map[string]struct{})

			dir.Build.GetRules(func(rule please.Rule) {
				kind := rule.Kind()

				switch kind {
				case "go_binary", "go_library", "go_test":
					rulesByKind[kind] = append(rulesByKind[kind], rule)
				}
			})

			if dir.Ok && dir.Rewrite && dir.Gopkg != nil {
				gopkg := dir.Gopkg

				if len(gopkg.GoFiles) > 0 {
					var kind string

					switch gopkg.Name {
					case "main":
						kind = "go_binary"
					default:
						kind = "go_library"
					}

					rules := rulesByKind[kind]
					name := filepath.Base(dir.Path)
					if len(rules) == 0 && dir.Build.GetRule(name) == nil {
						rule := this.please.NewRule(kind, name)
						rulesByKind[kind] = []please.Rule{rule}
						isGeneratedRule[name] = struct{}{}
					}
				}

				if len(gopkg.TestGoFiles)+len(gopkg.XTestGoFiles) > 0 {
					kind := "go_test"
					rules := rulesByKind[kind]
					name := "test"
					if len(rules) == 0 && dir.Build.GetRule(name) == nil {
						rule := this.please.NewRule(kind, name)
						rulesByKind[kind] = []please.Rule{rule}
						isGeneratedRule[name] = struct{}{}
					}
				}

			Rules:
				for _, kind := range []string{"go_binary", "go_library", "go_test"} {
					rules := rulesByKind[kind]

					for _, rule := range rules {
						for _, comment := range rule.Comment().Before {
							token := strings.TrimSpace(comment.Token)

							if strings.EqualFold(token, "# wollemi:keep") {
								continue
							}
						}

						var external bool
						var includePattern string
						var excludePattern string

						goFiles = goFiles[:0]
						imports = imports[:0]

						ruleName := rule.Name()

						_, isGeneratedRule := isGeneratedRule[ruleName]

						switch kind {
						case "go_binary", "go_library":
							goFiles = gopkg.GoFiles
							imports = gopkg.Imports
							includePattern = "*.go"
							excludePattern = "*_test.go"
						case "go_test":
							includePattern = "*_test.go"
							goFiles = gopkg.XTestGoFiles
							imports = gopkg.XTestImports
							external = len(goFiles) > 0

							if !external {
								includePattern = "*.go"
								goFiles = append(gopkg.GoFiles, gopkg.TestGoFiles...)
								imports = append(gopkg.Imports, gopkg.TestImports...)
							}
						}

						if config.ExplicitSources.IsTrue() {
							switch kind {
							case "go_test", "go_library", "go_binary":
								srcs := make([]string, 0, len(goFiles))

								for _, filename := range goFiles {
									info := dir.Files[filename]
									if info.Mode()&os.ModeSymlink == 0 { // is not symlink
										srcs = append(srcs, filename)
									}
								}

								if !isGeneratedRule {
									for _, target := range rule.AttrStrings("srcs") {
										switch {
										case strings.HasPrefix(target, ":"):
										case strings.HasPrefix(target, "//"):
										default:
											continue
										}

										srcs = append(srcs, target)
									}
								}

								rule.SetAttr("srcs", please.Strings(srcs...))
							}
						} else if !isGeneratedRule {
							goFiles = goFilesFromExpr(rule.Attr("srcs"), dir)
						}

						if len(goFiles) == 0 {
							log.WithField("build_rule", ruleName).
								WithField("reason", "no source files").
								Warn("removed")

							dir.Build.DelRule(ruleName)

							continue
						}

						log := log.WithField("build_rule", ruleName)

						if isGeneratedRule {
							include = include[:0]
							exclude = exclude[:0]
							targets = targets[:0]

							if !external {
								exclude = append(exclude, gopkg.IgnoredGoFiles...)
							}

							if includePattern != "" {
								include = append(include, includePattern)
							}

							if excludePattern != "" {
								exclude = append(exclude, excludePattern)
							}

							var srcLen int

							for _, filename := range goFiles {
								relpath := filepath.Join(path, filename)

								log := log.WithField("file", filename)

								target, _ := getTarget(config, relpath, true)
								if target == "" {
									info := dir.Files[filename]
									if info.Mode()&os.ModeSymlink == 0 { // is not symlink
										srcLen++

										if includePattern != "" {
											ok, err := filepath.Match(includePattern, filename)
											if err != nil {
												log.WithError(err).
													WithField("pattern", includePattern).
													Warn("could not match include pattern")

												continue
											}

											if ok {
												continue
											}

											if excludePattern != "" {
												ok, err := filepath.Match(excludePattern, filename)
												if err != nil {
													log.WithError(err).
														WithField("pattern", excludePattern).
														Warn("could not match exclude pattern")

													continue
												}

												if ok {
													continue
												}
											}
										}

										include = append(include, filename)
									}
								} else {
									targetPath, name := split(target)
									if targetPath == path {
										target = ":" + name
									}

									exclude = append(exclude, filename)
									targets = append(targets, target)
								}
							}

							if srcLen == 0 {
								continue
							}

							rule.SetAttr("srcs", please.Glob(include, exclude, targets...))
						}

						deps = deps[:0]
						unresolved = unresolved[:0]
						dedup := make(map[string]struct{}, len(imports))

						if !isGeneratedRule {
							if kind == "go_test" {
								// Allow internal go_test rule to depend on go_library rule even
								// though the go code does not. In the case of please this will
								// just make the pre-compiled go library code available in the
								// test.
								for _, dep := range rule.AttrStrings("deps") {
									path, name := split(dep)
									if path == "" || path == dir.Path {
										rule := dir.Build.GetRule(name)
										if rule != nil && rule.Kind() == "go_library" {
											dep = ":" + name

											if _, ok := dedup[dep]; !ok {
												dedup[dep] = struct{}{}
												deps = append(deps, dep)
											}
										}
									}
								}
							}

							// Since this is a pre-existing rule we will attempt to determine
							// the go imports using the srcs defined by the rule instead of the
							// go package.
							imports = imports[:0]

							for _, name := range goFiles {
								fileImports, ok := gopkg.GoFileImports[name]
								if !ok {
									// Attempt to recover by finding the missing go file imports
									// in plz-out/gen since the file could have been generated by
									// another rule.
									path := filepath.Join("plz-out/gen", dir.Path)
									gopkg, err := this.golang.ImportDir(path, []string{name})
									if err != nil {
										log.WithError(err).
											WithField("file", name).
											Error("could not parse required src file")

										continue Rules
									}

									fileImports, _ = gopkg.GoFileImports[name]
								}

								for _, path := range fileImports {
									var found bool

									for _, have := range imports {
										if path == have {
											found = true
											break
										}
									}

									if !found {
										imports = append(imports, path)
									}
								}
							}
						}

						for _, godep := range imports {
							if isGoroot[godep] {
								continue
							}

							target, _ := getTarget(config, godep, false)
							if target == "" {
								if !config.AllowUnresolvedDependency.IsTrue() {
									log.WithField("go_import", godep).
										Error("could not resolve go import")

									unresolved = append(unresolved, godep)
								}

								continue
							}

							targetPath, name := split(target)
							if targetPath == path {
								target = ":" + name
							}

							if _, ok := dedup[target]; !ok {
								dedup[target] = struct{}{}
								deps = append(deps, target)
							}
						}

						// skip rewrite because we have unresolved dependencies
						if len(unresolved) > 0 {
							continue
						}

						if isGeneratedRule {
							switch rule.Kind() {
							case "go_test":
								if external {
									rule.SetAttr("external", &please.Ident{Name: "True"})
								} else {
									rule.DelAttr("external")
								}
							case "go_binary", "go_library":
								visibility := config.DefaultVisibility

								if visibility == "" {
									visibility = "PUBLIC"

									for _, root := range paths {
										dir, name := split(root)
										if name == "..." && dir != "" {
											if path == dir || strings.HasPrefix(path, dir+"/") {
												visibility = fmt.Sprintf("//%s/...", dir)
												break
											}
										}
									}
								}

								rule.SetAttr("visibility", please.Strings(visibility))
							}
						}

						sort.Slice(deps, func(i, j int) bool {
							iPath, iName := split(deps[i])
							jPath, jName := split(deps[j])

							if iPath == jPath {
								return iName < jName
							}

							return iPath < jPath
						})

						rule.SetAttr("deps", please.Strings(deps...))

						if isGeneratedRule {
							dir.Build.SetRule(rule)
						}
					}
				}
			}

			if err := this.please.Write(dir.Build); err != nil {
				log.WithError(err).Warn("could not write")
			}
		})
	}

	gen.Close()
	get.Close()

	return nil
}

func goFilesFromExpr(expr please.Expr, dir *Directory) []string {
	var goFiles []string

	switch expr := expr.(type) {
	case *please.BinaryExpr:
		if expr.Op == "+" {
			goFiles = append(goFiles, goFilesFromExpr(expr.X, dir)...)
			goFiles = append(goFiles, goFilesFromExpr(expr.Y, dir)...)
		}
	case *please.ListExpr:
		for _, entry := range expr.List {
			switch s := entry.(type) {
			case *please.StringExpr:
				if strings.HasPrefix(s.Value, ":") || strings.HasPrefix(s.Value, "//") {
					path, name := split(s.Value)

					if path == "" || path == dir.Path {
						rule := dir.Build.GetRule(name)
						if rule != nil {
							goFiles = append(goFiles, goFilesFromExpr(rule.Unwrap(), dir)...)
						}
					}
				} else {
					goFiles = append(goFiles, s.Value)
				}
			}
		}
	case *please.CallExpr:
		var kind string

		switch x := expr.X.(type) {
		case *please.Ident:
			kind = x.Name
		}

		switch kind {
		case "genrule":
			outs := please.Attr(expr, "outs")
			goFiles = append(goFiles, goFilesFromExpr(outs, dir)...)
		case "go_copy":
			if name := please.AttrString(expr, "name"); name != "" {
				goFiles = append(goFiles, name+".cp.go")
			}
		case "glob":
			var include []string
			var exclude []string

			argLen := len(expr.List)

			if argLen >= 1 {
				switch expr := expr.List[0].(type) {
				case *please.ListExpr:
					for _, entry := range expr.List {
						switch s := entry.(type) {
						case *please.StringExpr:
							include = append(include, s.Value)
						}
					}
				}
			}

			if argLen > 1 {
				switch assign := expr.List[1].(type) {
				case *please.AssignExpr:
					switch lhs := assign.LHS.(type) {
					case *please.Ident:
						if lhs.Name == "exclude" {
							switch rhs := assign.RHS.(type) {
							case *please.ListExpr:
								for _, entry := range rhs.List {
									switch s := entry.(type) {
									case *please.StringExpr:
										exclude = append(exclude, s.Value)
									}
								}
							}
						}
					}
				}
			}

			for _, xs := range [][]string{
				dir.Gopkg.GoFiles,
				dir.Gopkg.TestGoFiles,
				dir.Gopkg.XTestGoFiles,
			} {
				for _, x := range xs {
					if isMatched(x, include) && !isMatched(x, exclude) {
						goFiles = append(goFiles, x)
					}
				}
			}
		case "filegroup":
			srcs := please.Attr(expr, "srcs")
			goFiles = append(goFiles, goFilesFromExpr(srcs, dir)...)
		}
	}

	return goFiles
}

func isMatched(s string, patterns []string) bool {
	for _, pattern := range patterns {
		if s == pattern {
			return true
		}

		ok, err := filepath.Match(pattern, s)
		if err != nil {
			panic(err)
		}

		if ok {
			return true
		}
	}

	return false
}
