package wollemi_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tcncloud/wollemi/testdata/expect"
	"github.com/tcncloud/wollemi/testdata/please"
)

func TestService_RulesUnused(t *testing.T) {
	NewServiceSuite(t).TestService_RulesUnused()
}

func (t *ServiceSuite) TestService_RulesUnused() {
	type T = ServiceSuite

	const (
		gopkg = "github.com/wollemi_test"
		gosrc = "/go/src"
	)

	t.It("can list unused build rules", func(t *T) {
		t.MockRulesUnused()

		wollemi := t.New(gosrc, gopkg)

		var (
			prune        bool
			kinds        []string
			paths        []string
			excludePaths []string
		)

		wollemi.RulesUnused(prune, kinds, paths, excludePaths)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"kind":  "filegroup",
				"msg":   "unused",
				"name":  "files",
				"path":  "app",
			},
			map[string]interface{}{
				"level": "info",
				"kind":  "go_get",
				"msg":   "unused",
				"name":  "viper",
				"path":  "third_party/go/github.com/spf13",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})

	t.It("can prune unused build rules", func(t *T) {
		t.MockRulesUnused()

		t.please.EXPECT().Write(any).Times(2).Do(func(have please.File) {
			path := have.GetPath()

			var want *please.BuildFile

			switch path {
			case "app/BUILD.plz":
				want = &please.BuildFile{
					Path: path,
					Stmt: []please.Expr{
						please.NewCallExpr("app", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				}
			case "third_party/go/github.com/spf13/BUILD.plz":
				want = &please.BuildFile{
					Path: path,
					Stmt: []please.Expr{
						please.NewCallExpr("go_get", []please.Expr{
							please.NewAssignExpr("=", "name", "cobra"),
							please.NewAssignExpr("=", "get", "github.com/spf13/cobra"),
							please.NewAssignExpr("=", "revision", "v1.0.0"),
							please.NewAssignExpr("=", "deps", []string{":pflag"}),
						}),
						please.NewCallExpr("go_get", []please.Expr{
							please.NewAssignExpr("=", "name", "pflag"),
							please.NewAssignExpr("=", "get", "github.com/spf13/pflag"),
							please.NewAssignExpr("=", "revision", "v1.0.5"),
						}),
					},
				}
			default:
				t.Errorf("unexpected call to please write: %s", path)
			}

			expect.Equal(t, want, have)
		})

		wollemi := t.New(gosrc, gopkg)

		var (
			prune        bool = true
			kinds        []string
			paths        []string
			excludePaths []string
		)

		wollemi.RulesUnused(prune, kinds, paths, excludePaths)
	})
}

func (t *ServiceSuite) MockRulesUnused() {
	graph := &please.Graph{
		Packages: map[string]*please.Package{
			"app": &please.Package{
				Targets: map[string]*please.Target{
					"files": &please.Target{},
					"app": &please.Target{
						Deps: []string{
							"//third_party/go/github.com/spf13:cobra",
						},
					},
				},
			},
			"third_party/go/github.com/spf13": &please.Package{
				Targets: map[string]*please.Target{
					"cobra": &please.Target{
						Deps: []string{
							"//third_party/go/github.com/spf13:pflag",
						},
					},
					"viper": &please.Target{},
					"pflag": &please.Target{},
				},
			},
		},
	}

	t.please.EXPECT().Graph().Return(graph, nil)

	t.filesystem.EXPECT().ReadDir(any).AnyTimes().
		DoAndReturn(func(path string) (infos []os.FileInfo, err error) {
			switch path {
			case "app":
				infos = []os.FileInfo{
					&FileInfo{
						FileName: "main.go",
						FileMode: os.FileMode(420),
					},
					&FileInfo{
						FileName: "main_test.go",
						FileMode: os.FileMode(420),
					},
					&FileInfo{
						FileName: "BUILD.plz",
						FileMode: os.FileMode(420),
					},
				}
			case "third_party/go/github.com/spf13":
				infos = []os.FileInfo{
					&FileInfo{
						FileName: "BUILD.plz",
						FileMode: os.FileMode(420),
					},
				}
			default:
				t.Errorf("unexpected call to filesystem read dir: %s", path)
				err = os.ErrNotExist
			}

			return infos, err
		})

	t.filesystem.EXPECT().ReadAll(any, any).AnyTimes().
		DoAndReturn(func(buf *bytes.Buffer, path string) error {
			switch path {
			case "app/BUILD.plz":
			case "third_party/go/github.com/spf13/BUILD.plz":
			default:
				t.Errorf("unexpected call to filesystem read all: %s", path)
			}

			buf.Reset()
			buf.WriteString(path)

			return nil
		})

	t.please.EXPECT().Parse(any, any).AnyTimes().
		DoAndReturn(func(path string, data []byte) (please.File, error) {
			assert.Equal(t, string(data), path)

			var file *please.BuildFile

			switch path {
			case "app/BUILD.plz":
				file = &please.BuildFile{
					Path: path,
					Stmt: []please.Expr{
						please.NewCallExpr("filegroup", []please.Expr{
							please.NewAssignExpr("=", "name", "files"),
						}),
						please.NewCallExpr("app", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				}
			case "third_party/go/github.com/spf13/BUILD.plz":
				file = &please.BuildFile{
					Path: path,
					Stmt: []please.Expr{
						please.NewCallExpr("go_get", []please.Expr{
							please.NewAssignExpr("=", "name", "cobra"),
							please.NewAssignExpr("=", "get", "github.com/spf13/cobra"),
							please.NewAssignExpr("=", "revision", "v1.0.0"),
							please.NewAssignExpr("=", "deps", []string{":pflag"}),
						}),
						please.NewCallExpr("go_get", []please.Expr{
							please.NewAssignExpr("=", "name", "pflag"),
							please.NewAssignExpr("=", "get", "github.com/spf13/pflag"),
							please.NewAssignExpr("=", "revision", "v1.0.5"),
						}),
						please.NewCallExpr("go_get", []please.Expr{
							please.NewAssignExpr("=", "name", "viper"),
							please.NewAssignExpr("=", "get", "github.com/spf13/pflag"),
							please.NewAssignExpr("=", "revision", "v1.6.3"),
						}),
					},
				}
			default:
				t.Errorf("unexpected call to please parse: %s", path)
			}

			return file, nil
		})
}
