package wollemi

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/tcncloud/wollemi/ports/please"
)

func (this *Service) RulesUnused(prune bool, kinds, paths, exclude []string) error {
	graph, err := this.please.Graph()
	if err != nil {
		return err
	}

	paths = this.normalizePaths(paths)

	if len(kinds) == 0 {
		kinds = []string{
			"export_file",
			"filegroup",
			"genrule",
			"go_get",
			"go_get_with_sources",
			"go_mock",
			"go_library",
			"grpc_library",
			"pip_library",
		}
	}

	deps := make(map[string]struct{})
	rules := make(map[string]struct{})
	revdeps := make(map[string][]string)

	for path, pkg := range graph.Packages {
		for name, target := range pkg.Targets {
			rule := path + ":" + name

			rules[rule] = struct{}{}

			for _, dep := range target.Deps {
				dep = strings.TrimPrefix(dep, "//")
				revdeps[dep] = append(revdeps[dep], rule)
				deps[dep] = struct{}{}
			}
		}
	}

	collect := make(chan *Directory, 1000)
	parse := make(chan *Directory, 1000)

	var parser sync.WaitGroup

	for i := 0; i < runtime.NumCPU()-1; i++ {
		parser.Add(1)
		go func() {
			defer parser.Done()

			buf := bytes.NewBuffer(nil)

			for dir := range parse {
				dir.Build = (func() please.File {
					log := this.log.WithField("path", dir.Path)

					buildPath, err := this.FindBuildFile(dir.Path)
					if err != nil {
						log.WithError(err).Warn("could not find build file")
						return nil
					}

					if err := this.filesystem.ReadAll(buf, buildPath); err != nil {
						log.WithError(err).Warn("could not read build file")
						return nil
					}

					file, err := this.please.Parse(buildPath, buf.Bytes())
					if err != nil {
						log.WithError(err).Warn("could not parse build file")
						return nil
					}

					return file
				}())

				collect <- dir
			}
		}()
	}

	dirs := make(map[string]*Directory)

	var collector sync.WaitGroup

	collector.Add(1)
	go func() {
		defer collector.Done()
		for dir := range collect {
			dirs[dir.Path] = dir
		}
	}()

	parsed := make(map[string]struct{})
	unused := make(map[string][]string)

	for rule, _ := range rules {
		if !inRunPath(rule, paths...) {
			continue
		}

		if len(revdeps[rule]) != 0 {
			continue
		}

		target := please.Split(rule)

		unused[target.Path] = append(unused[target.Path], target.Name)

		if _, ok := parsed[target.Path]; ok {
			continue
		}

		nonBlockingSend(parse, &Directory{Path: target.Path})
		parsed[target.Path] = struct{}{}
	}

	close(parse)
	parser.Wait()

	close(collect)
	collector.Wait()

	isIncludeKind := make(map[string]struct{}, len(kinds))
	for _, name := range kinds {
		isIncludeKind[name] = struct{}{}
	}

Unused:
	for path, names := range unused {
		log := this.log.WithField("path", path)
		dir, ok := dirs[path]
		if !ok || dir == nil {
			continue
		}

		switch dir.Build.(type) {
		case nil:
			continue
		}

		for _, prefix := range exclude {
			if strings.HasPrefix(path, prefix) {
				continue Unused
			}
		}

		var pruned int

	UnusedNames:
		for _, name := range names {
			rule := dir.Build.GetRule(name)
			if rule == nil {
				continue
			}

			kind := rule.Kind()

			if _, ok := isIncludeKind[kind]; !ok {
				continue
			}

			switch kind {
			case "grpc_library":
				for _, x := range []string{"proto", "go", "java", "py", "ts"} {
					if len(revdeps[fmt.Sprintf("%s:_%s#%s", path, name, x)]) != 0 {
						continue UnusedNames
					}
				}
			case "pip_library":
				if len(revdeps[path+":_"+name+"#wheel"]) != 0 {
					continue
				}
			}

			if prune {
				if dir.Build.DelRule(name) {
					pruned++
				}
			} else {
				log.WithField("name", name).
					WithField("kind", kind).
					Info("unused")
			}
		}

		if pruned > 0 {
			if err := this.please.Write(dir.Build); err != nil {
				log.WithError(err).Warn("could not write")
			}
		}
	}

	return nil
}
