package bazel

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bazelbuild/buildtools/build"

	"github.com/tcncloud/wollemi/ports/logging"
	"github.com/tcncloud/wollemi/ports/please"
)

func NewBuilder(log logging.Logger, ctl please.Ctl, filesystem please.Filesystem) *Builder {
	return &Builder{
		Ctl:        ctl,
		filesystem: filesystem,
		log:        log,
	}
}

type Builder struct {
	please.Ctl
	filesystem please.Filesystem
	log        logging.Logger
}

func (this *Builder) Parse(path string, data []byte) (please.File, error) {
	file, err := build.ParseBuild(path, data)
	if err != nil {
		return nil, err
	}

	return NewFile(data, file), nil
}

func (this *Builder) NewFile(path string) please.File {
	return NewFile(nil, &build.File{
		Type: build.TypeBuild,
		Path: path,
	})
}

func (this *Builder) NewRule(kind, name string) please.Rule {
	return NewRule(&build.CallExpr{
		X: &build.Ident{Name: kind},
		List: []build.Expr{
			&build.AssignExpr{
				LHS: &build.Ident{Name: "name"},
				Op:  "=",
				RHS: &build.StringExpr{Value: name},
			},
		},
	})
}

func (this *Builder) Write(file please.File) error {
	switch file := file.(type) {
	case *File:
		for _, stmt := range file.Stmt {
			switch call := stmt.(type) {
			case *build.CallExpr:
				for _, attr := range call.List {
					switch attr := attr.(type) {
					case *build.AssignExpr:
						switch lhs := attr.LHS.(type) {
						case *build.Ident:
							if lhs.Name != "srcs" {
								continue
							}
						}

						switch rhs := attr.RHS.(type) {
						case *build.CallExpr:
							switch x := rhs.X.(type) {
							case *build.Ident:
								if x.Name == "glob" {
									start, end := rhs.Span()
									if start.Line == end.Line {
										rhs.ForceCompact = true
									}
								}
							}
						}
					}
				}
			}
		}

		log := this.log.WithField("path", filepath.Join("/", filepath.Dir(file.Path)))

		if file.IsEmpty() {
			err := this.filesystem.Remove(file.Path)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("could not remove build file: %v", err)
				}
			} else {
				log.Info("deleted")
			}

			for path := filepath.Dir(file.Path); true; path = filepath.Dir(path) {
				infos, err := this.filesystem.ReadDir(path)
				if err != nil {
					return err
				}

				if len(infos) != 0 {
					break
				}

				if err := this.filesystem.Remove(path); err != nil {
					return err
				}
			}

			return nil
		}

		data := build.Format(file.Unwrap())
		if !bytes.Equal(file.Data, data) {
			err := this.filesystem.WriteFile(file.Path, data, os.FileMode(0644))
			if err != nil {
				return fmt.Errorf("could not write build file: %v", err)
			}

			log.Info("modified")
		}

		return nil
	default:
		panic(fmt.Errorf("file not created by package"))
	}
}
