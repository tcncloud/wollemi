package wollemi_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tcncloud/wollemi/ports/golang"
	"github.com/tcncloud/wollemi/testdata/please"
)

func TestService_SymlinkList(t *testing.T) {
	NewServiceSuite(t).TestService_SymlinkList()
}

func (t *ServiceSuite) TestService_SymlinkList() {
	type T = ServiceSuite

	const (
		gopkg = "github.com/wollemi_test"
		gosrc = "/go/src"
	)

	t.It("can list all project symlinks", func(t *T) {
		data := t.GoSymlinkTestData()

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					if err := walkFn(path, info, nil); err != nil {
						return err
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		wollemi := t.New(gosrc, gopkg)

		var (
			name    = "*"
			broken  bool
			prune   bool
			exclude []string
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/service.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/service.pb.go",
			},
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/entities.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/entities.pb.go",
			},
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/mock/mock.mg.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/mock/mock.mg.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})

	t.It("can list only broken symlinks", func(t *T) {
		data := t.GoSymlinkTestData()

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					if err := walkFn(path, info, nil); err != nil {
						return err
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		t.filesystem.EXPECT().Stat(any).AnyTimes().
			DoAndReturn(func(path string) (interface{}, error) {
				prefix := filepath.Join(gosrc, gopkg, "plz-out/gen") + "/"

				assert.Regexp(t, "^"+prefix, path)

				switch strings.TrimPrefix(path, prefix) {
				case "app/protos/entities.pb.go":
					return nil, os.ErrNotExist
				case "app/protos/mock/mock.mg.go":
					return nil, os.ErrNotExist
				}

				info := &FileInfo{
					FileName: filepath.Base(path),
					FileMode: os.FileMode(420),
				}

				return info, nil
			})

		wollemi := t.New(gosrc, gopkg)

		var (
			name         = "*"
			broken  bool = true
			prune   bool
			exclude []string
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/entities.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/entities.pb.go",
			},
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/mock/mock.mg.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/mock/mock.mg.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})

	t.It("can list symlinks matching name", func(t *T) {
		data := t.GoSymlinkTestData()

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					if err := walkFn(path, info, nil); err != nil {
						return err
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		wollemi := t.New(gosrc, gopkg)

		var (
			name    = "*.mg.go"
			broken  bool
			prune   bool
			exclude []string
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/mock/mock.mg.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/mock/mock.mg.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})

	t.It("can prune matched symlinks", func(t *T) {
		data := t.GoSymlinkTestData()

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					if err := walkFn(path, info, nil); err != nil {
						return err
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		t.filesystem.EXPECT().Stat(any).AnyTimes().
			DoAndReturn(func(path string) (interface{}, error) {
				prefix := filepath.Join(gosrc, gopkg, "plz-out/gen") + "/"

				assert.Regexp(t, "^"+prefix, path)

				switch strings.TrimPrefix(path, prefix) {
				case "app/protos/service.pb.go":
					return nil, os.ErrNotExist
				}

				info := &FileInfo{
					FileName: filepath.Base(path),
					FileMode: os.FileMode(420),
				}

				return info, nil
			})

		t.filesystem.EXPECT().Remove("app/protos/service.pb.go")

		wollemi := t.New(gosrc, gopkg)

		var (
			name         = "*.pb.go"
			broken  bool = true
			prune   bool = true
			exclude []string
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/service.pb.go",
				"msg":   "symlink deleted",
				"path":  "plz-out/gen/app/protos/service.pb.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})

	t.It("can exclude listing symlinks by path prefix", func(t *T) {
		data := t.GoSymlinkTestData()

		skipDir := make(map[string]struct{})

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

			Walk:
				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					for prefix, _ := range skipDir {
						if strings.HasPrefix(path, prefix) {
							continue Walk
						}
					}

					err := walkFn(path, info, nil)
					if err == filepath.SkipDir {
						skipDir[path] = struct{}{}
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		wollemi := t.New(gosrc, gopkg)

		var (
			name    = "*"
			broken  bool
			prune   bool
			exclude []string = []string{
				"app/protos/mock",
			}
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/service.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/service.pb.go",
			},
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/entities.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/entities.pb.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})
}

func (t *ServiceSuite) GoSymlinkTestData() *GoFormatTestData {
	data := &GoFormatTestData{
		Gosrc: gosrc,
		Gopkg: gopkg,
		Paths: []string{"app/..."},
		ImportDir: map[string]*golang.Package{
			"app/protos": &golang.Package{
				GoFiles: []string{
					"service.pb.go",
					"entities.pb.go",
				},
				GoFileImports: map[string][]string{
					"service.pb.go": []string{
						"github.com/golang/protobuf/proto",
						"google.golang.org/grpc",
						"google.golang.org/grpc/codes",
						"google.golang.org/grpc/status",
					},
					"entities.pb.go": []string{
						"github.com/golang/protobuf/proto",
					},
				},
			},
			"app/protos/mock": &golang.Package{
				Name: "mock_protos",
				GoFiles: []string{
					"mock.mg.go",
				},
				GoFileImports: map[string][]string{
					"mock.mg.go": []string{
						"github.com/golang/mock/gomock",
						"github.com/example/app/protos",
						"google.golang.org/grpc",
						"google.golang.org/grpc/metadata",
					},
				},
			},
			"app/server": &golang.Package{
				GoFiles: []string{
					"server.go",
				},
				XTestGoFiles: []string{
					"server_test.go",
				},
				GoFileImports: map[string][]string{
					"server_test.go": []string{
						"github.com/golang/mock/gomock",
						"github.com/golang/protobuf/proto/ptypes/wrappers",
						"github.com/stretchr/testify/assert",
						"github.com/stretchr/testify/require",
						"github.com/example/app/protos/mock",
						"testing",
					},
					"server.go": []string{
						"database/sql",
						"encoding/json",
						"github.com/golang/protobuf/proto/ptypes/wrappers",
						"github.com/example/app/protos",
						"google.golang.org/grpc",
						"google.golang.org/grpc/credentials",
						"strconv",
						"strings",
					},
				},
			},
			"app": &golang.Package{
				Name: "main",
				GoFiles: []string{
					"main.go",
				},
				TestGoFiles: []string{
					"main_test.go",
				},
				GoFileImports: map[string][]string{
					"main_test.go": []string{
						"github.com/golang/mock/gomock",
						"github.com/stretchr/testify/assert",
						"github.com/stretchr/testify/require",
						"testing",
					},
					"main.go": []string{
						"fmt",
						"github.com/spf13/cobra",
						"github.com/example/app/server",
					},
				},
			},
		},
		Lstat: map[string]*FileInfo{
			"app/protos/service.pb.go": &FileInfo{
				FileName: "service.pb.go",
				FileMode: os.FileMode(134218221),
			},
			"app/protos/entities.pb.go": &FileInfo{
				FileName: "entities.pb.go",
				FileMode: os.FileMode(134218221),
			},
			"app/protos/mock/mock.mg.go": &FileInfo{
				FileName: "mock.mg.go",
				FileMode: os.FileMode(134218221),
			},
		},
		Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
			"app/BUILD.plz": &please.BuildFile{
				Stmt: []please.Expr{
					please.NewCallExpr("go_binary", []please.Expr{
						please.NewAssignExpr("=", "name", "app"),
						please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
						please.NewAssignExpr("=", "visibility", []string{"//app/..."}),
						please.NewAssignExpr("=", "deps", []string{
							"//app/server",
							"//third_party/go/github.com/spf13:cobra",
						}),
					}),
					please.NewCallExpr("go_test", []please.Expr{
						please.NewAssignExpr("=", "name", "test"),
						please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"})),
						please.NewAssignExpr("=", "deps", []string{
							"//app/server",
							"//third_party/go/github.com/golang:mock",
							"//third_party/go/github.com/spf13:cobra",
							"//third_party/go/github.com/stretchr:testify",
						}),
					}),
				},
			},
			"app/server/BUILD.plz": &please.BuildFile{
				Stmt: []please.Expr{
					please.NewCallExpr("go_library", []please.Expr{
						please.NewAssignExpr("=", "name", "server"),
						please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
						please.NewAssignExpr("=", "visibility", []string{"//app/..."}),
						please.NewAssignExpr("=", "deps", []string{
							"//app/protos",
							"//third_party/go/github.com/golang:protobuf",
							"//third_party/go/google.golang.org:grpc",
						}),
					}),
					please.NewCallExpr("go_test", []please.Expr{
						please.NewAssignExpr("=", "name", "test"),
						please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*_test.go"})),
						please.NewAssignExpr("=", "external", true),
						please.NewAssignExpr("=", "deps", []string{
							"//app/protos:mock",
							"//third_party/go/github.com/golang:mock",
							"//third_party/go/github.com/golang:protobuf",
							"//third_party/go/github.com/stretchr:testify",
						}),
					}),
				},
			},
			"app/protos/BUILD.plz": &please.BuildFile{
				Stmt: []please.Expr{
					please.NewCallExpr("grpc_library", []please.Expr{
						please.NewAssignExpr("=", "name", "protos"),
						please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.proto"})),
						please.NewAssignExpr("=", "protoc_flags", []string{
							"-I third_party/proto",
							"-I .",
						}),
						please.NewAssignExpr("=", "visibility", []string{"//app/..."}),
						please.NewAssignExpr("=", "labels", []string{"link:app/protos"}),
					}),
					please.NewCallExpr("go_mock", []please.Expr{
						please.NewAssignExpr("=", "name", "mock"),
						please.NewAssignExpr("=", "package", "github.com/example/app/protos"),
						please.NewAssignExpr("=", "visibility", []string{"//..."}),
						please.NewAssignExpr("=", "deps", []string{
							":protos",
							"//third_party/go/github.com/golang:mock",
							"//third_party/go/google.golang.org:grpc",
						}),
					}),
				},
			},
		}),
		Readlink: map[string]string{
			"app/protos/service.pb.go":   "plz-out/gen/app/protos/service.pb.go",
			"app/protos/entities.pb.go":  "plz-out/gen/app/protos/entities.pb.go",
			"app/protos/mock/mock.mg.go": "plz-out/gen/app/protos/mock/mock.mg.go",
		},
	}

	data.Prepare()

	return data
}
