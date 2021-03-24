package wollemi

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/tcncloud/wollemi/ports/logging"
	"github.com/tcncloud/wollemi/ports/please"
	"github.com/tcncloud/wollemi/ports/wollemi"
)

const (
	BUILD_FILE = "BUILD.plz"
)

func (this *Service) Format(config wollemi.Config, paths []string) error {
	config.Gofmt.Rewrite = wollemi.Bool(false)

	return this.GoFormat(config, paths)
}

func (this *Service) isInternal(path string) bool {
	if this.gopkg != "" && strings.HasPrefix(path, this.gopkg) {
		return true
	}

	for _, path := range []string{path, filepath.Dir(path)} {
		if path == "." {
			continue
		}

		for _, prefix := range []string{"", "plz-out/gen"} {
			info, err := this.filesystem.Stat(filepath.Join(prefix, path))

			if err != nil && !os.IsNotExist(err) {
				this.log.WithField("path", path).Warnf("could not stat: %v", err)
			}

			if err == nil && info.IsDir() {
				return true
			}
		}
	}

	return false
}

func (this *Service) GoFormat(config wollemi.Config, paths []string) error {
	this.gofmt.paths = this.normalizePaths(paths)
	this.config = config

	log := this.log.WithField("rewrite", this.config.Gofmt.GetRewrite())
	for i, path := range this.gofmt.paths {
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

	this.gofmt.isGoroot = make(map[string]bool)

	for i := 0; i < runtime.NumCPU()-1; i++ {
		go func() {
			var buf bytes.Buffer

			for dir := range parse {
				dir.InRunPath = inRunPath(dir.Path, this.gofmt.paths...)
				dir.Rewrite = this.config.Gofmt.GetRewrite()

				nonBlockingSend(collect, this.ParseDir(&buf, dir))
			}
		}()
	}

	if err := this.ReadDirs(walk, this.gofmt.paths...); err != nil {
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

						goroot, ok := this.gofmt.isGoroot[godep]
						if !ok {
							if prefix = this.isInternal(godep); prefix {
								goroot = false
							} else {
								goroot = this.golang.IsGoroot(godep)
							}

							this.gofmt.isGoroot[godep] = goroot
						}

						if goroot {
							continue
						}

						path := godep

						if prefix || this.isInternal(path) {
							path = strings.TrimPrefix(path, this.gopkg+"/")
						} else {
							path = filepath.Join("third_party/go", path)
						}

						if _, ok := external[godep]; ok {
							continue
						}

						if inRunPath(path, this.gofmt.paths...) {
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
				default:
					name := rule.AttrString("name")

					target, path := dir.Path, dir.Path
					if path == "." {
						target = fmt.Sprintf(":%s", name)
					} else if filepath.Base(path) != name {
						path = filepath.Join(path, name)
						target += ":" + name
					}

					importPath := rule.AttrString("import_path")
					kind := rule.Kind()

					switch {
					case importPath != "":
						external[importPath] = append(external[path], "//"+target)
					case strings.HasPrefix(path, "third_party/go/"):
					case kind != "go_test":
						internal[filepath.Join(this.gopkg, path)] = "//" + target

						if rule.Kind() == "go_copy" {
							genfiles[path+".cp.go"] = target
						}
					}
				}
			})
		}
	}

	close(parse)

	get := NewChanFunc(1, 0)
	gen := NewChanFunc(runtime.NumCPU()-1, 0)

	this.gofmt.getTarget = (func() func(wollemi.Config, string, bool) (string, string) {
		var inner func(wollemi.Config, string, bool, int) (string, string)

		inner = func(config wollemi.Config, path string, isFile bool, depth int) (string, string) {
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

			if depth == 0 {
				if target, ok := internal[path]; ok {
					return target, path
				}

				if _, ok := directories[path]; ok {
					return fmt.Sprintf("//%s", path), path
				}

				if this.isInternal(path) {
					if dir, ok := directories[filepath.Dir(path)]; ok {
						name := filepath.Base(path)

						if dir.Build != nil {
							if rule := dir.Build.GetRule(name); rule != nil {
								return fmt.Sprintf("//%s:%s", dir.Path, name), dir.Path
							}
						}
					}

					if path == this.gopkg {
						path = fmt.Sprintf(":%s", filepath.Base(this.wd))
					} else {
						path = strings.TrimPrefix(path, this.gopkg+"/")
					}

					return fmt.Sprintf("//%s", path), path
				}
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

			return inner(config, path, isFile, depth+1)
		}

		return func(config wollemi.Config, path string, isFile bool) (string, string) {
			var target string

			get.RunBlock(func() {
				target, path = inner(config, path, isFile, 0)
			})

			return target, path
		}
	}())

	for path, dir := range directories {
		if !dir.InRunPath {
			continue
		}

		log := this.log.WithField("path", filepath.Join("/", path))
		dir := dir

		gen.Run(func() {
			this.formatDirectory(log, dir)

			if err := this.please.Write(dir.Build); err != nil {
				log.WithError(err).Warn("could not write")
			}
		})
	}

	gen.Close()
	get.Close()

	return nil
}

func (this *Service) getRuleDeps(files []string, config wollemi.Config, dir *Directory) ([]string, []string, error) {
	var imports, resolved, unresolved []string

	for _, name := range files {
		fileImports, ok := dir.Gopkg.GoFileImports[name]
		if !ok {
			// Attempt to recover by finding the missing go file imports
			// in plz-out/gen since the file could have been generated by
			// another rule.
			path := filepath.Join("plz-out/gen", dir.Path)
			gopkg, err := this.golang.ImportDir(path, []string{name})
			if err != nil {
				return nil, nil, fmt.Errorf("could not parse src file: %s", name)
			}

			fileImports, _ = gopkg.GoFileImports[name]
		}

		for _, path := range fileImports {
			imports = appendUniqString(imports, path)
		}

		for _, path := range imports {
			if this.gofmt.isGoroot[path] {
				continue
			}

			targetPath, _ := this.gofmt.getTarget(config, path, false)
			if targetPath == "" {
				unresolved = append(unresolved, path)
				continue
			}

			target := please.Split(targetPath)
			resolved = appendUniqString(resolved, target.Rel(dir.Path))
		}
	}

	please.SortDeps(resolved)

	return resolved, unresolved, nil
}

func (this *Service) getVisibility(config wollemi.Config, path string) string {
	if config.DefaultVisibility != "" {
		return config.DefaultVisibility
	}

	for _, root := range this.gofmt.paths {
		target := please.Split(root)

		if target.Path == "." {
			return "//..."
		}

		if target.Name == "..." && target.Path != "" {
			if path == target.Path || strings.HasPrefix(path, target.Path+"/") {
				return fmt.Sprintf("//%s/...", target.Path)
			}
		}
	}

	return "PUBLIC"
}

func (this *Service) getRuleSrcs(dir *Directory, config wollemi.Config, srcFiles []string) []string {
	srcs := make([]string, 0, len(srcFiles))

	for _, name := range srcFiles {
		relpath := filepath.Join(dir.Path, name)

		targetPath, _ := this.gofmt.getTarget(config, relpath, true)
		if targetPath == "" {
			info, ok := dir.Files[name]
			if !ok {
				continue
			}

			if info.Mode()&os.ModeSymlink == 0 { // is not symlink
				srcs = append(srcs, name)
			}
		} else {
			target := please.Split(targetPath)

			srcs = append(srcs, target.Rel(dir.Path))
		}
	}

	return srcs
}

func (this *Service) formatDirectory(log logging.Logger, dir *Directory) {
	if !dir.Ok || !dir.Rewrite || dir.Gopkg == nil {
		return
	}

	config := this.filesystem.Config(dir.Path).Merge(this.config)

	filesByRule := make(map[string][]string)
	rulesByFile := make(map[string][]string)

	dir.Build.GetRules(func(rule please.Rule) {
		files := getSrcFilesFromExpr(rule.Attr("srcs"), dir)

		for i := 0; i < len(files); i++ {
			name := files[i]

			switch {
			case filepath.Ext(name) != ".go":
				// Ignore non golang source files.
			case strings.Contains(name, "/"):
				// Ignore golang source files outside of this directory.
			default:
				if _, ok := dir.Files[name]; !ok {
					// This golang source file does not exist so it should be stripped
					// from this rule's source list.

					if i+1 < len(files) {
						copy(files[i:], files[i+1:])
					}

					files = files[:len(files)-1]
					i--
				}
			}
		}

		filesByRule[rule.Name()] = files

		for _, name := range files {
			rulesByFile[name] = appendUniqString(rulesByFile[name], rule.Name())
		}
	})

	// ---------------------------------------------------------------------------
	// Manage existing go rules in this directory.

	dir.Build.GetRules(func(rule please.Rule) {
		for _, comment := range rule.Comment().Before {
			token := strings.TrimSpace(comment.Token)

			if strings.EqualFold(token, "# wollemi:keep") {
				return // TODO: write unit test for this.
			}
		}

		log := log.WithField("build_rule", rule.Name())

		var pkgFiles []string
		var newFiles []string
		var external bool

		kind := config.Gofmt.GetMapped(rule.Kind())

		if !inStrings(config.Gofmt.GetManage(), kind) {
			return
		}

		switch kind {
		case "go_binary", "go_library":
			pkgFiles = dir.Gopkg.GoFiles
		case "go_test":
			pkgFiles = dir.Gopkg.XTestGoFiles
			external = true

			if len(pkgFiles) == 0 {
				pkgFiles = append(dir.Gopkg.GoFiles, dir.Gopkg.TestGoFiles...)
				external = false
			}
		}

		var ambiguous bool

		srcFiles := filesByRule[rule.Name()]

		for _, name := range pkgFiles {
			if inStrings(srcFiles, name) {
				continue
			}

			if rulesByFile[name] == nil {
				newFiles = append(newFiles, name)
				continue
			}

			if kind == "go_test" && inStrings(dir.Gopkg.GoFiles, name) {
				newFiles = append(newFiles, name)
				continue
			}

			ambiguous = true
			break
		}

		if len(newFiles) > 0 && !ambiguous {
			for _, name := range newFiles {
				rulesByFile[name] = appendUniqString(rulesByFile[name], rule.Name())
			}

			srcFiles = append(srcFiles, newFiles...)
			newFiles = nil

			sort.Strings(srcFiles)
		}

		if kind == "go_test" && !external {
			if len(srcFiles) == len(dir.Gopkg.GoFiles) {
				sort.Strings(dir.Gopkg.GoFiles)
				sort.Strings(srcFiles)

				tail := len(srcFiles) - 1
				for i := 0; i <= tail; i++ {
					if srcFiles[i] != dir.Gopkg.GoFiles[i] {
						break
					}

					if i == tail {
						// This internal go_test does not include any actual test source
						// files therefore we are marking it to be deleted.
						srcFiles = nil
					}
				}
			}
		}

		if len(srcFiles) == 0 {
			log.WithField("build_rule", rule.Name()).
				WithField("reason", "no source files").
				Warn("removed")

			dir.Build.DelRule(rule.Name())
			return
		}

		var internal []string

		if kind == "go_test" && !external {
			// Allow internal go_test rule to depend on go_library rule even
			// though the go code does not. In the case of please this will
			// just make the pre-compiled go library code available in the
			// test.
			for _, dep := range rule.AttrStrings("deps") {
				target := please.Split(dep)

				if target.Path == "" || target.Path == dir.Path {
					rule := dir.Build.GetRule(target.Name)

					if rule != nil && rule.Kind() == "go_library" {
						internal = append(internal, fmt.Sprintf(":%s", target.Name))
						srcFiles = deleteStrings(srcFiles, filesByRule[target.Name]...)
					}
				}
			}
		}

		srcs := this.getRuleSrcs(dir, config, srcFiles)

		resolved, unresolved, err := this.getRuleDeps(srcFiles, config, dir)
		if err != nil {
			log.WithError(err).Warn("could not get deps")
			return
		}

		resolved = append(internal, resolved...)

		if len(unresolved) > 0 && !config.AllowUnresolvedDependency.IsTrue() {
			for _, path := range unresolved {
				log.WithField("go_import", path).Error("could not resolve go import")
			}

			return
		}

		_, isExplicitSources := rule.Attr("srcs").(*please.ListExpr)

		if isExplicitSources || config.ExplicitSources.IsTrue() {
			rule.SetAttr("srcs", please.Strings(srcs...))
		}

		if len(resolved) > 0 {
			rule.SetAttr("deps", please.Strings(resolved...))
		}
	})

	// ---------------------------------------------------------------------------
	// Create missing go rules in this directory.

	for _, kind := range []string{"go_library", "go_test"} {
		var pkgFiles []string
		var include []string
		var exclude []string
		var rule please.Rule
		var external bool

		switch kind {
		case "go_library":
			pkgFiles = dir.Gopkg.GoFiles

			if dir.Gopkg.Name == "main" {
				kind = "go_binary"
			}

			path := filepath.Join(this.wd, dir.Path)
			kind := config.Gofmt.GetMapped(kind)
			rule = this.please.NewRule(kind, filepath.Base(path))
			include, exclude = []string{"*.go"}, []string{"*_test.go"}
		case "go_test":
			if len(dir.Gopkg.XTestGoFiles)+len(dir.Gopkg.TestGoFiles) == 0 {
				continue
			}

			pkgFiles = dir.Gopkg.XTestGoFiles
			kind := config.Gofmt.GetMapped(kind)
			rule = this.please.NewRule(kind, "test")
			include = []string{"*_test.go"}
			external = true

			if len(pkgFiles) == 0 {
				pkgFiles = append(dir.Gopkg.GoFiles, dir.Gopkg.TestGoFiles...)
				include = []string{"*.go"}
				external = false
			}
		}

		if !inStrings(config.Gofmt.GetCreate(), kind) {
			continue
		}

		log := log.WithField("build_rule", rule.Name())

		srcFiles := make([]string, 0, len(pkgFiles))

		var ambiguous bool

		for _, name := range pkgFiles {
			if rulesByFile[name] == nil {
				srcFiles = append(srcFiles, name)
				continue
			}

			if kind == "go_test" && inStrings(dir.Gopkg.GoFiles, name) {
				srcFiles = append(srcFiles, name)
				continue
			}

			ambiguous = true
			break
		}

		if ambiguous {
			continue
		}

		for _, name := range srcFiles {
			rulesByFile[name] = appendUniqString(rulesByFile[name], rule.Name())
		}

		srcs := this.getRuleSrcs(dir, config, srcFiles)

		if len(srcs) == 0 {
			continue
		}

		if config.ExplicitSources.IsTrue() {
			rule.SetAttr("srcs", please.Strings(srcs...))
		} else {
			if !external {
				exclude = append(exclude, dir.Gopkg.IgnoredGoFiles...)
			}

			rule.SetAttr("srcs", please.Glob(include, exclude))
		}

		if external {
			rule.SetAttr("external", &please.Ident{Name: "True"})
		}

		if kind == "go_binary" || kind == "go_library" {
			visibility := this.getVisibility(config, dir.Path)

			rule.SetAttr("visibility", please.Strings(visibility))
		}

		resolved, unresolved, err := this.getRuleDeps(srcFiles, config, dir)
		if err != nil {
			log.WithError(err).Warn("could not get deps")
			return
		}

		if len(unresolved) > 0 && !config.AllowUnresolvedDependency.IsTrue() {
			for _, path := range unresolved {
				log.WithField("go_import", path).Error("could not resolve go import")
			}

			return
		}

		if len(resolved) > 0 {
			rule.SetAttr("deps", please.Strings(resolved...))
		}

		dir.Build.SetRule(rule)
	}
}

func getSrcFilesFromExpr(expr please.Expr, dir *Directory) []string {
	var srcFiles []string

	switch expr := expr.(type) {
	case *please.ListExpr:
		for _, entry := range expr.List {
			switch s := entry.(type) {
			case *please.StringExpr:
				if strings.HasPrefix(s.Value, ":") || strings.HasPrefix(s.Value, "//") {
					target := please.Split(s.Value)

					if target.Path == "" || target.Path == dir.Path {
						rule := dir.Build.GetRule(target.Name)
						if rule != nil {
							srcFiles = append(srcFiles, getSrcFilesFromExpr(rule.Unwrap(), dir)...)
						}
					}
				} else {
					srcFiles = append(srcFiles, s.Value)
				}
			}
		}
	case *please.BinaryExpr:
		if expr.Op == "+" {
			srcFiles = append(srcFiles, getSrcFilesFromExpr(expr.X, dir)...)
			srcFiles = append(srcFiles, getSrcFilesFromExpr(expr.Y, dir)...)
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
			srcFiles = append(srcFiles, getSrcFilesFromExpr(outs, dir)...)
		case "go_copy":
			if name := please.AttrString(expr, "name"); name != "" {
				srcFiles = append(srcFiles, name+".cp.go")
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
						srcFiles = append(srcFiles, x)
					}
				}
			}
		case "filegroup":
			srcs := please.Attr(expr, "srcs")
			srcFiles = append(srcFiles, getSrcFilesFromExpr(srcs, dir)...)
		}
	}

	return srcFiles
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

func inStrings(from []string, value string) bool {
	return indexStrings(from, value) >= 0
}

func appendUniqString(dest []string, from ...string) []string {
	for _, s := range from {
		if !inStrings(dest, s) {
			dest = append(dest, s)
		}
	}

	return dest
}

func indexStrings(from []string, value string) int {
	for i, have := range from {
		if value == have {
			return i
		}
	}

	return -1
}

func deleteStrings(from []string, values ...string) []string {
	for _, s := range values {
		if i := indexStrings(from, s); i >= 0 {
			if i+1 < len(from) {
				copy(from[i:], from[i+1:])
			}

			from = from[:len(from)-1]
		}
	}

	return from
}
