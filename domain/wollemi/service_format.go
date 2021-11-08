package wollemi

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/tcncloud/wollemi/domain/optional"
	"github.com/tcncloud/wollemi/ports/logging"
	"github.com/tcncloud/wollemi/ports/please"
	"github.com/tcncloud/wollemi/ports/wollemi"
)

const (
	BUILD_FILE = "BUILD.plz"
)

func newGoFormat(paths []string) *goFormat {
	return &goFormat{
		resolveLimiter: NewChanFunc(1, 0),
		isGoroot:       map[string]bool{},
		paths:          paths,
		directories:    map[string]*Directory{},
		external:       map[string][]string{},
		internal:       map[string]string{},
		genfiles:       map[string]string{},
	}
}

// goFormat contains all the state for formatting go rules i.e. the import path to target mapping, and the rules in the
// targeted Please packages.
type goFormat struct {
	// isGoroot contains a cache of import paths that are part of the Go SDK
	isGoroot map[string]bool

	// paths are the normalised paths we were asked to format
	paths []string

	// resolveLimiter is used to control the concurrency on resolving targets.
	resolveLimiter *ChanFunc

	// directories represent the set of Please packages we have parsed
	directories map[string]*Directory

	// external is a map of third party imports to build targets
	external map[string][]string

	// internal is a map of this projects imports paths to targets
	internal map[string]string

	// genfiles contain a map of generated files added via go_copy() to their build targets
	genfiles map[string]string
}

// getTarget gets the target for an import path or generated file
func (this *Service) getTarget(config wollemi.Config, p string, isFile bool) string {
	var target string
	this.goFormat.resolveLimiter.RunBlock(func() {
		target, _ = this.getTargetInternal(config, p, isFile, 0)
	})
	return target
}

func (this *Service) getTargetInternal(config wollemi.Config, path string, isFile bool, depth int) (string, string) {
	if target, ok := config.KnownDependency[path]; ok {
		return target, path
	}

	if isFile && filepath.Ext(path) == ".go" {
		if target, ok := this.goFormat.genfiles[path]; ok {
			return target, path
		} else {
			return "", path
		}
	}

	if depth == 0 {
		if target, ok := this.goFormat.internal[path]; ok {
			return target, path
		}

		if _, ok := this.goFormat.directories[path]; ok {
			return fmt.Sprintf("//%s", path), path
		}

		if this.isInternal(path) {
			if dir, ok := this.goFormat.directories[filepath.Dir(path)]; ok {
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

	targets, ok := this.goFormat.external[path]

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

	return this.getTargetInternal(config, path, isFile, depth+1)
}

// parsePaths will start parsing the Please packages to be formatted. It populates the directories map on the goFormat
// struct. These can later be formatted with formatDirs().
func (this *Service) parsePaths() error {
	collect := make(chan *Directory, 1000)
	parse := make(chan *Directory, 1000)
	walk := make(chan *Directory, 1000)

	for i := 0; i < runtime.NumCPU()-1; i++ {
		go func() {
			var buf bytes.Buffer

			for dir := range parse {
				dir.InRunPath = inRunPath(dir.Path, this.goFormat.paths...)
				dir.Rewrite = this.config.Gofmt.GetRewrite()

				nonBlockingSend(collect, this.ParseDir(&buf, dir))
			}
		}()
	}

	if err := this.ReadDirs(walk, this.goFormat.paths...); err != nil {
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

			this.goFormat.directories[dir.Path] = dir

			if dir.Gopkg != nil {
				for _, imports := range [][]string{
					dir.Gopkg.Imports,
					dir.Gopkg.TestImports,
					dir.Gopkg.XTestImports,
				} {
				Imports:
					for _, godep := range imports {
						goroot, ok := this.goFormat.isGoroot[godep]

						if !ok {
							goroot = this.golang.IsGoroot(godep)

							this.goFormat.isGoroot[godep] = goroot
						}

						if goroot {
							continue
						}

						path := godep

						if this.isInternal(path) {
							path = strings.TrimPrefix(path, this.gopkg+"/")
						} else {
							path = filepath.Join("third_party/go", path)
						}

						if _, ok := this.goFormat.external[godep]; ok {
							continue
						}

						if inRunPath(path, this.goFormat.paths...) {
							continue
						}

						chunks := strings.Split(path, "/")

						for i := len(chunks); i > 0; i-- {
							path := filepath.Join(chunks[0:i]...)

							if _, ok := delegated[path]; ok {
								continue Imports
							}

							if _, ok := this.goFormat.directories[path]; ok {
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
				case "go_module":
					module := rule.AttrString("module")
					target := &please.Target{
						Name: rule.AttrString("name"),
						Path: dir.Path,
					}

					for _, install := range rule.AttrStrings("install") {
						install = strings.TrimSuffix(install, "/...")
						path := module

						if install != "." {
							path = filepath.Join(path, install)
						}

						this.goFormat.external[path] = append(this.goFormat.external[path], target.String())
					}
				case "go_get", "go_get_with_sources":
					get := strings.TrimSuffix(rule.AttrString("get"), "/...")
					if get == "" {
						if install := rule.AttrStrings("install"); len(install) > 0 {
							sort.Strings(install)
							get = strings.TrimSuffix(install[0], "/...")
						}
					}

					target := &please.Target{
						Name: rule.AttrString("name"),
						Path: dir.Path,
					}

					if rule.Kind() == "go_get_with_sources" {
						get = rule.AttrStrings("outs")[0]
					}

					if get != "" && rule.AttrLiteral("binary") != "True" {
						this.goFormat.external[get] = append(this.goFormat.external[get], target.String())
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

					target = strings.TrimSuffix(target, ":")

					importPath := rule.AttrString("import_path")
					kind := rule.Kind()

					switch {
					case importPath != "":
						this.goFormat.external[importPath] = append(this.goFormat.external[path], "//"+target)
					case strings.HasPrefix(path, "third_party/go/"):
					case kind != "go_test":
						this.goFormat.internal[filepath.Join(this.gopkg, path)] = "//" + target

						if kind == "go_copy" {
							this.goFormat.genfiles[path+".cp.go"] = target
						}
					}
				}
			})
		}
	}

	close(parse)
	return nil
}

// formatDirs updated the BUILD file in the directories that were in the original paths
func (this *Service) formatDirs() {
	limiter := NewChanFunc(runtime.NumCPU()-1, 0)
	defer limiter.Close()

	for path, dir := range this.goFormat.directories {
		if !dir.InRunPath {
			continue
		}

		log := this.log.WithField("path", filepath.Join("/", path))
		dir := dir

		limiter.Run(func() {
			this.formatDir(log, dir)

			if err := this.please.Write(dir.Build); err != nil {
				log.WithError(err).Warn("could not write")
			}
		})
	}
}

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
	this.goFormat = newGoFormat(this.normalizePaths(paths))
	defer this.goFormat.resolveLimiter.Close()

	this.config = config

	log := this.log.WithField("rewrite", this.config.Gofmt.GetRewrite())

	for i, path := range this.goFormat.paths {
		log = log.WithField(fmt.Sprintf("path[%d]", i), path)
	}

	log.Debug("running gofmt")

	if err := this.parsePaths(); err != nil {
		return err
	}
	this.formatDirs()

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
			if this.goFormat.isGoroot[path] {
				continue
			}

			targetPath := this.getTarget(config, path, false)
			if targetPath == "" {
				unresolved = append(unresolved, path)
				continue
			}

			target := please.Split(targetPath)
			resolved = appendUniqString(resolved, target.Rel(dir.Path))
		}
	}

	return resolved, unresolved, nil
}

func (this *Service) getVisibility(config wollemi.Config, path string) string {
	if config.DefaultVisibility != "" {
		return config.DefaultVisibility
	}

	for _, root := range this.goFormat.paths {
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

		targetPath := this.getTarget(config, relpath, true)
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

func (this *Service) formatDir(log logging.Logger, dir *Directory) {
	if !dir.Ok || !dir.Rewrite || dir.Gopkg == nil {
		return
	}

	config := this.filesystem.Config(dir.Path).Merge(this.config)

	log.WithFields(logging.Fields{
		"config":    config,
		"directory": dir,
	}).Debug("formatting")

	consumer := &fileConsumer{
		Rules: make(map[string][]string),
		Files: make(map[string][]string),
		Dir:   dir,
	}

	var managed []please.Rule

	dir.Build.GetRules(func(rule please.Rule) {
		kind := config.Gofmt.GetMapped(rule.Kind())

		if inStrings(config.Gofmt.GetManage(), kind) {
			managed = append(managed, rule)
		}

		consumer.Update(rule)
	})

	sortManagedRules(config, managed)

	// ---------------------------------------------------------------------------
	// Manage existing go rules in this directory.

ManageRules:
	for {
		for i, rule := range managed {
			for _, comment := range rule.Comment().Before {
				token := strings.TrimSpace(comment.Token)

				if strings.EqualFold(token, "# wollemi:keep") {
					return // TODO: write unit test for this.
				}
			}

			log := log.WithFields(logging.Fields{
				"rule":    rule.Name(),
				"process": "manage",
			})

			var pkgFiles []string
			var deps []string

			kind := config.Gofmt.GetMapped(rule.Kind())

			external := rule.AttrLiteral("external") == "True"

			switch kind {
			case "go_binary", "go_library":
				pkgFiles = dir.Gopkg.GoFiles
			case "go_test":
				if external {
					pkgFiles = dir.Gopkg.XTestGoFiles
				} else {
					pkgFiles = dir.Gopkg.TestGoFiles
				}
			}

			srcFiles := consumer.Files[rule.Name()]

			// -----------------------------------------------------------------------
			// Include missing golang package source files unless one or more source
			// files are being consumed by another rule. Allow exceptions when
			// internal go_test rules consume non test package sources.

			var ambiguous bool

			for i := 0; !ambiguous && i < len(pkgFiles); i++ {
				file := pkgFiles[i]

				switch {
				case len(consumer.Rules[file]) == 0:
				case inStrings(srcFiles, file):
				case inStrings(dir.Gopkg.GoFiles, file):
					if kind == "go_test" {
						continue
					}

					for _, name := range consumer.Rules[file] {
						other := dir.Build.GetRule(name)
						if other == nil {
							continue
						}

						if config.Gofmt.GetMapped(other.Kind()) != "go_test" {
							ambiguous = true
							break
						}
					}
				default:
					ambiguous = true
				}
			}

			if !ambiguous {
				srcFiles = appendUniqString(srcFiles, pkgFiles...)
				sort.Strings(srcFiles)
			}

			// -----------------------------------------------------------------------
			// Mark external go_test rule to be removed when it includes no test
			// source files.

			if kind == "go_test" && !external {
				tail := len(srcFiles) - 1

				for i := 0; i <= tail; i++ {
					if strings.HasSuffix(srcFiles[i], "_test.go") {
						break
					} else if i == tail {
						srcFiles = nil // delete
					}
				}
			}

			// -----------------------------------------------------------------------
			// Remove managed build rule when it includes zero source files.

			if len(srcFiles) == 0 {
				log.WithField("reason", "no source files").Warn("removed")

				rule.DelAttr("srcs")
				consumer.Update(rule)

				dir.Build.DelRule(rule.Name())

				managed = deleteRulesIndex(managed, i)

				continue ManageRules
			}

			// -----------------------------------------------------------------------

			if kind == "go_test" && !external {
				// Allow internal go_test rule to depend on go_library rule even
				// though the go code does not. In the case of please this will
				// just make the pre-compiled go library code available in the
				// test.
				for _, dep := range rule.AttrStrings("deps") {
					target := please.Split(dep)

					if target.Path == "" || target.Path == dir.Path {
						rule := dir.Build.GetRule(target.Name)

						if rule != nil && config.Gofmt.GetMapped(rule.Kind()) == "go_library" {
							deps = append(deps, fmt.Sprintf(":%s", target.Name))
							srcFiles, _ = deleteStrings(srcFiles, consumer.Files[target.Name]...)
						}

						if strings.HasSuffix(target.Name, "#lib") {
							name := target.Name

							name = strings.TrimSuffix(name, "#lib")
							name = strings.TrimPrefix(name, "_")

							rule := dir.Build.GetRule(name)

							if rule != nil && config.Gofmt.GetMapped(rule.Kind()) == "go_binary" {
								deps = append(deps, fmt.Sprintf(":%s", target.Name))
								srcFiles, _ = deleteStrings(srcFiles, consumer.Files[name]...)
							}
						}
					}
				}
			}

			_, isExplicitSources := rule.Attr("srcs").(*please.ListExpr)

			if kind == "go_test" && !isExplicitSources {
				exclude := dir.Gopkg.XTestGoFiles

				if external {
					exclude = dir.Gopkg.TestGoFiles
				}

				srcFiles, isExplicitSources = deleteStrings(srcFiles, exclude...)
			}

			if external {
				if inStrings(dir.Gopkg.TestGoFiles, srcFiles...) {
					rule.DelAttr("external")
				}
			} else {
				if inStrings(dir.Gopkg.XTestGoFiles, srcFiles...) {
					rule.SetAttr("external", &please.Ident{Name: "True"})
				}
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

			if isExplicitSources || config.ExplicitSources.IsTrue() {
				srcs := this.getRuleSrcs(dir, config, srcFiles)
				rule.SetAttr("srcs", please.Strings(srcs...))
			}

			resolved = append(deps, resolved...)

			if len(resolved) > 0 {
				please.SortDeps(resolved)
				rule.SetAttr("deps", please.Strings(resolved...))
			} else {
				rule.DelAttr("deps")
			}

			log.Debug("managed")

			consumer.Update(rule)
		}

		break
	}

	// ---------------------------------------------------------------------------
	// Create missing go rules in this directory.

CreateRules:
	for _, x := range []struct {
		Kind     string
		External bool
	}{
		{Kind: "go_library"},
		{Kind: "go_test", External: false},
		{Kind: "go_test", External: true},
	} {
		var pkgFiles []string
		var include []string
		var exclude []string
		var rule please.Rule
		var deps []string

		config := config

		switch {
		case x.Kind == "go_library":
			pkgFiles = dir.Gopkg.GoFiles

			if dir.Gopkg.Name == "main" {
				x.Kind = "go_binary"
			}

			rule = this.please.NewRule(
				config.Gofmt.GetMapped(x.Kind),
				filepath.Base(filepath.Join(this.wd, dir.Path)),
			)

			include, exclude = []string{"*.go"}, []string{"*_test.go"}
		case x.Kind == "go_test" && x.External == false:
			if len(dir.Gopkg.TestGoFiles) == 0 {
				continue // No source files so nothing to be done.
			}

			pkgFiles = dir.Gopkg.TestGoFiles

			if dir.Gopkg.Name == "main" {
				pkgFiles = append(dir.Gopkg.GoFiles, dir.Gopkg.TestGoFiles...)
				include = []string{"*.go"}
			} else {
				// Attempt to get sources through existing go_library rule.
				if rule := consumer.GetRule(dir.Gopkg.GoFiles...); rule != nil {
					if config.Gofmt.GetMapped(rule.Kind()) == "go_library" {
						// Since there exists exactly one go_library rule which consumes all
						// non-test source files for this go package it will be included as
						// a dependency for this internal go_test.

						deps = append(deps, ":"+rule.Name())
						pkgFiles = dir.Gopkg.TestGoFiles
						include = []string{"*_test.go"}
					}
				}
			}

			exclude = dir.Gopkg.XTestGoFiles
			name := "test"

			if len(dir.Gopkg.XTestGoFiles) > 0 {
				config.ExplicitSources = optional.BoolValue(true)
				name = "internal_test"
			}

			rule = this.please.NewRule(config.Gofmt.GetMapped(x.Kind), name)
		case x.Kind == "go_test" && x.External == true:
			if len(dir.Gopkg.XTestGoFiles) == 0 {
				continue // No sources files so nothing to be done.
			}

			pkgFiles = dir.Gopkg.XTestGoFiles
			include = []string{"*_test.go"}
			exclude = dir.Gopkg.TestGoFiles
			name := "test"

			if len(dir.Gopkg.TestGoFiles) > 0 {
				config.ExplicitSources = optional.BoolValue(true)
				name = "external_test"
			}

			rule = this.please.NewRule(config.Gofmt.GetMapped(x.Kind), name)
		}

		if !inStrings(config.Gofmt.GetCreate(), x.Kind) {
			continue
		}

		exclude = append(exclude, dir.Gopkg.IgnoredGoFiles...)

		log := log.WithFields(logging.Fields{
			"rule":    rule.Name(),
			"process": "create",
		})

		for _, file := range pkgFiles {
			switch {
			case len(consumer.Rules[file]) == 0:
			case x.Kind == "go_test" && inStrings(dir.Gopkg.GoFiles, file):
			default:
				log.WithFields(logging.Fields{
					"rules":  consumer.Rules[file],
					"reason": "ambiguous",
					"file":   file,
				}).Debug("skipped")

				continue CreateRules
			}
		}

		srcs := this.getRuleSrcs(dir, config, pkgFiles)
		if len(srcs) == 0 {
			log.WithField("reason", "no sources").Debug("skipped")

			continue
		}

		if config.ExplicitSources.IsTrue() {
			rule.SetAttr("srcs", please.Strings(srcs...))
		} else {
			rule.SetAttr("srcs", please.Glob(include, exclude))
		}

		if x.External {
			rule.SetAttr("external", &please.Ident{Name: "True"})
		}

		if x.Kind == "go_binary" || x.Kind == "go_library" {
			visibility := this.getVisibility(config, dir.Path)

			rule.SetAttr("visibility", please.Strings(visibility))
		}

		resolved, unresolved, err := this.getRuleDeps(pkgFiles, config, dir)
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

		resolved = append(deps, resolved...)

		if len(resolved) > 0 {
			please.SortDeps(resolved)
			rule.SetAttr("deps", please.Strings(resolved...))
		}

		consumer.Update(rule)

		dir.Build.SetRule(rule)

		log.Debug("created")
	}

	for _, files := range [][]string{
		dir.Gopkg.GoFiles,
		dir.Gopkg.TestGoFiles,
		dir.Gopkg.XTestGoFiles,
	} {
		for _, file := range files {
			if _, ok := consumer.Rules[file]; !ok {
				info, ok := dir.Files[file]
				if !ok {
					continue
				}

				if info.Mode()&os.ModeSymlink == 0 { // is not symlink
					log.WithField("file", file).Debug("unsourced")
				}
			}
		}
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

func inStrings(from []string, values ...string) bool {
	for _, value := range values {
		if indexStrings(from, value) < 0 {
			return false
		}
	}

	return true
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

func deleteStrings(from []string, values ...string) ([]string, bool) {
	var deleted bool

	for _, s := range values {
		if i := indexStrings(from, s); i >= 0 {
			deleted = true

			if i+1 < len(from) {
				copy(from[i:], from[i+1:])
			}

			from = from[:len(from)-1]
		}
	}

	return from, deleted
}

func sortManagedRules(config wollemi.Config, rules []please.Rule) {
	sort.Slice(rules, func(i, j int) bool {
		irule := rules[i]
		jrule := rules[j]

		ikind := config.Gofmt.GetMapped(irule.Kind())
		jkind := config.Gofmt.GetMapped(jrule.Kind())

		switch {
		case ikind == jkind:
			return irule.Name() < jrule.Name()
		case ikind == "go_library":
			return true
		case ikind == "go_binary" && jkind != "go_library":
			return true
		case ikind == "go_test" && jkind != "go_binary" && jkind != "go_library":
			return true
		default:
			return ikind < jkind
		}
	})
}

func deleteRulesIndex(rules []please.Rule, i int) []please.Rule {
	size := len(rules)

	if i+1 < size {
		copy(rules[i:], rules[i+1:])
	}

	return rules[:size-1]
}

type fileConsumer struct {
	Rules map[string][]string
	Files map[string][]string
	Dir   *Directory
}

func (fc *fileConsumer) GetRule(files ...string) please.Rule {
	var names []string

	for _, file := range files {
		names = appendUniqString(names, fc.Rules[file]...)
		if len(names) > 1 {
			return nil
		}
	}

	var rule please.Rule

	if len(names) == 1 {
		rule = fc.Dir.Build.GetRule(names[0])
	}

	return rule
}

func (fc *fileConsumer) Update(rule please.Rule) {
	delete(fc.Files, rule.Name())

	for file, rules := range fc.Rules {
		fc.Rules[file], _ = deleteStrings(rules, rule.Name())
	}

	files := getSrcFilesFromExpr(rule.Attr("srcs"), fc.Dir)

	for i := 0; i < len(files); i++ {
		name := files[i]

		switch {
		case filepath.Ext(name) != ".go":
			// Ignore non golang source files.
		case strings.Contains(name, "/"):
		// Ignore golang source files outside of this directory.
		case strings.HasPrefix(name, "//"), strings.HasPrefix(name, ":"):
		// Ignore please rule targets
		default:
			if _, ok := fc.Dir.Files[name]; !ok {
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

	fc.Files[rule.Name()] = files

	for _, file := range files {
		fc.Rules[file] = appendUniqString(fc.Rules[file], rule.Name())
	}
}
