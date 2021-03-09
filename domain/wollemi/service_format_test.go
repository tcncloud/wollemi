package wollemi_test

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/domain/optional"
	"github.com/tcncloud/wollemi/ports/filesystem"
	"github.com/tcncloud/wollemi/ports/golang"
	"github.com/tcncloud/wollemi/testdata/expect"
	"github.com/tcncloud/wollemi/testdata/please"
)

func TestService_GoFormat(t *testing.T) {
	NewServiceSuite(t).TestService_GoFormat()
}

func (t *ServiceSuite) TestService_GoFormat() {
	type T = ServiceSuite

	const (
		gosrc   = "/go/src"
		gopkg   = "github.com/wollemi_test"
		rewrite = true
	)

	t.It("can generate missing go_binary with go_test", func(t *T) {
		data := t.GoFormatTestData()

		file := data.Parse["app/BUILD.plz"]
		require.NotNil(t, file)

		data.Stat[file.Path] = nil
		data.Lstat[file.Path] = nil
		data.Parse[file.Path] = nil
		data.Write[file.Path] = file

		rule := file.GetRule("app")
		require.NotNil(t, rule)
		rule.SetAttr("visibility", please.NewListExpr("PUBLIC"))

		t.MockGoFormat(data)

		wollemi := t.New(gosrc, gopkg)

		require.NoError(t, wollemi.GoFormat(rewrite, []string{"app"}))
	})

	t.It("can generate missing go_library with x go_test", func(t *T) {
		data := t.GoFormatTestData()

		file := data.Parse["app/server/BUILD.plz"]
		require.NotNil(t, file)

		data.Stat[file.Path] = nil
		data.Lstat[file.Path] = nil
		data.Parse[file.Path] = nil
		data.Write[file.Path] = file

		rule := file.GetRule("server")
		require.NotNil(t, rule)
		rule.SetAttr("visibility", please.NewListExpr("PUBLIC"))

		t.MockGoFormat(data)

		wollemi := t.New(gosrc, gopkg)

		require.NoError(t, wollemi.GoFormat(rewrite, []string{"app/server"}))
	})

	t.It("can recursively generate missing go rules", func(t *T) {
		data := t.GoFormatTestData()

		for _, path := range []string{
			"app/server/BUILD.plz",
			"app/BUILD.plz",
		} {
			file := data.Parse[path]
			require.NotNil(t, file)

			data.Stat[file.Path] = nil
			data.Lstat[file.Path] = nil
			data.Parse[file.Path] = nil
			data.Write[file.Path] = file
		}

		file := data.Parse["app/protos/BUILD.plz"]
		require.NotNil(t, file)

		data.Write[file.Path] = file

		// TODO: This is ok because the file will not actually get generated since
		// it is empty. However, it would be better if the implementation did not
		// attempt to write this build file at all.
		data.Parse["app/protos/mock/BUILD.plz"] = nil
		data.Write["app/protos/mock/BUILD.plz"] = &please.BuildFile{
			Path: "app/protos/mock/BUILD.plz",
		}

		t.MockGoFormat(data)

		wollemi := t.New(gosrc, gopkg)

		require.NoError(t, wollemi.GoFormat(rewrite, []string{"app/..."}))
	})

	t.It("can format existing go_library rule", func(t *T) {
		data := t.GoFormatTestData()

		want := data.Parse["app/server/BUILD.plz"]
		require.NotNil(t, want)

		have := (please.Copier{}).BuildFile(want)

		data.Write["app/server/BUILD.plz"] = want

		rule := have.GetRule("server")
		require.NotNil(t, rule)

		rule.SetAttr("deps", please.NewListExpr())

		t.MockGoFormat(data)

		wollemi := t.New(gosrc, gopkg)

		require.NoError(t, wollemi.GoFormat(rewrite, []string{"app/server"}))
	})

	t.It("supports multiple go_test rules", func(t *T) {
		data := t.GoFormatTestData()

		data.Config["app/example"] = &filesystem.Config{}

		data.ImportDir["app/example"] = &golang.Package{
			Name: "example",
			XTestGoFiles: []string{
				"one_test.go",
				"two_test.go",
			},
			XTestImports: []string{
				"github.com/spf13/pflag",
				"github.com/spf13/cobra",
				"testing",
			},
			GoFileImports: map[string][]string{
				"one_test.go": []string{
					"github.com/spf13/pflag",
					"testing",
				},
				"two_test.go": []string{
					"github.com/spf13/cobra",
					"testing",
				},
			},
		}

		data.Parse["app/example/BUILD.plz"] = &please.BuildFile{
			Path: "app/example/BUILD.plz",
			Stmt: []please.Expr{
				please.NewCallExpr("go_test", []please.Expr{
					please.NewAssignExpr("=", "name", "one_test"),
					please.NewAssignExpr("=", "srcs", please.NewListExpr("one_test.go")),
					please.NewAssignExpr("=", "external", true),
				}),
				please.NewCallExpr("go_test", []please.Expr{
					please.NewAssignExpr("=", "name", "two_test"),
					please.NewAssignExpr("=", "srcs", please.NewListExpr("two_test.go")),
					please.NewAssignExpr("=", "external", true),
				}),
			},
		}

		data.Write["app/example/BUILD.plz"] = &please.BuildFile{
			Path: "app/example/BUILD.plz",
			Stmt: []please.Expr{
				please.NewCallExpr("go_test", []please.Expr{
					please.NewAssignExpr("=", "name", "one_test"),
					please.NewAssignExpr("=", "srcs", please.NewListExpr("one_test.go")),
					please.NewAssignExpr("=", "external", true),
					please.NewAssignExpr("=", "deps", please.NewListExpr(
						"//third_party/go/github.com/spf13:pflag",
					)),
				}),
				please.NewCallExpr("go_test", []please.Expr{
					please.NewAssignExpr("=", "name", "two_test"),
					please.NewAssignExpr("=", "srcs", please.NewListExpr("two_test.go")),
					please.NewAssignExpr("=", "external", true),
					please.NewAssignExpr("=", "deps", please.NewListExpr(
						"//third_party/go/github.com/spf13:cobra",
					)),
				}),
			},
		}

		t.MockGoFormat(data)

		wollemi := t.New(gosrc, gopkg)

		require.NoError(t, wollemi.GoFormat(rewrite, []string{"app/example"}))
	})

	t.It("allows internal go_test to depend on go_library without go import", func(t *T) {
		data := t.GoFormatTestData()

		have := data.Parse["app/server/BUILD.plz"]
		require.NotNil(t, have)

		// -------------------------------------------------------------------------

		want := (please.Copier{}).BuildFile(have)

		rule := want.GetRule("test")
		require.NotNil(t, rule)

		deps := []interface{}{":server"}
		for _, dep := range rule.AttrStrings("deps") {
			deps = append(deps, dep)
		}

		rule.SetAttr("deps", please.NewListExpr(deps...))
		rule.DelAttr("external")

		data.Write["app/server/BUILD.plz"] = want

		// -------------------------------------------------------------------------

		rule = have.GetRule("test")
		require.NotNil(t, rule)

		rule.DelAttr("external")
		rule.SetAttr("deps", please.NewListExpr(":server"))

		t.MockGoFormat(data)

		wollemi := t.New(gosrc, gopkg)

		require.NoError(t, wollemi.GoFormat(rewrite, []string{"app/server"}))
	})

	t.It("can be configured to allow unresolved dependencies", func(t *T) {
		data := t.GoFormatTestData()

		for path, _ := range data.Stat {
			if strings.HasPrefix(path, "third_party/go/") {
				delete(data.Stat, path)
				delete(data.Lstat, path)
			}
		}

		for path, _ := range data.Parse {
			if strings.HasPrefix(path, "third_party/go/") {
				delete(data.Parse, path)
			}
		}

		data.Walk = make([]string, 0, len(data.Walk))
		for path, _ := range data.Lstat {
			data.Walk = append(data.Walk, path)
		}

		data.Config["app/server"] = &filesystem.Config{
			DefaultVisibility:         "//app/...",
			AllowUnresolvedDependency: optional.BoolValue(true),
		}

		file := data.Parse["app/server/BUILD.plz"]
		require.NotNil(t, file)

		data.Stat[file.Path] = nil
		data.Lstat[file.Path] = nil
		data.Parse[file.Path] = nil
		data.Write[file.Path] = file

		for _, name := range []string{"server", "test"} {
			rule := file.GetRule(name)
			have := rule.AttrStrings("deps")
			want := make([]interface{}, 0, len(have))

			for _, dep := range have {
				if !strings.HasPrefix(dep, "//third_party/go/") {
					want = append(want, dep)
				}
			}

			rule.SetAttr("deps", please.NewListExpr(want...))
		}

		t.MockGoFormat(data)

		wollemi := t.New(gosrc, gopkg)

		require.NoError(t, wollemi.GoFormat(rewrite, []string{"app/server"}))
	})

	t.It("can resolve dependencies by import_path", func(t *T) {
		data := t.GoFormatTestData()

		cobra := "third_party/go/github.com/spf13/cobra/BUILD.plz"
		data.Stat[cobra] = &FileInfo{
			FileName: "BUILD.plz",
			FileMode: os.FileMode(420),
		}
		data.Lstat[cobra] = data.Stat[cobra]

		data.Walk = make([]string, 0, len(data.Walk))
		for path, _ := range data.Lstat {
			data.Walk = append(data.Walk, path)
		}

		sort.Strings(data.Walk)

		data.Parse[cobra] = &please.BuildFile{
			Path: cobra,
			Stmt: []please.Expr{
				please.NewCallExpr("go_library", []please.Expr{
					please.NewAssignExpr("=", "name", "cobra"),
					please.NewAssignExpr("=", "import_path", "github.com/spf13/cobra"),
				}),
			},
		}

		file := data.Parse["app/BUILD.plz"]
		require.NotNil(t, file)

		data.Stat[file.Path] = nil
		data.Lstat[file.Path] = nil
		data.Parse[file.Path] = nil
		data.Write[file.Path] = file

		app := file.GetRule("app")
		require.NotNil(t, app)

		test := file.GetRule("test")
		require.NotNil(t, test)

		app.SetAttr("visibility", please.NewListExpr("PUBLIC"))
		app.SetAttr("deps", please.NewListExpr(
			"//app/server",
			"//third_party/go/github.com/spf13/cobra",
		))

		test.SetAttr("deps", please.NewListExpr(
			"//app/server",
			"//third_party/go/github.com/golang:mock",
			"//third_party/go/github.com/spf13/cobra",
			"//third_party/go/github.com/stretchr:testify",
		))

		t.MockGoFormat(data)

		wollemi := t.New(gosrc, gopkg)

		require.NoError(t, wollemi.GoFormat(rewrite, []string{"app"}))
	})
}

func (t *ServiceSuite) MockGoFormat(td *GoFormatTestData) {
	td.Build()

	t.golang.EXPECT().ImportDir(any, any).AnyTimes().
		DoAndReturn(func(path string, names []string) (*golang.Package, error) {
			gopkg, ok := td.ImportDir[path]
			if !ok {
				t.Errorf("unexpected call to golang import dir: %s", path)
			}

			return gopkg, nil
		})

	t.golang.EXPECT().IsGoroot(any).AnyTimes().
		DoAndReturn(func(path string) bool {
			goroot, ok := td.IsGoroot[path]
			if !ok {
				t.Errorf("unexpected call to golang is goroot: %s", path)
			}

			return goroot
		})

	t.please.EXPECT().Parse(any, any).AnyTimes().
		DoAndReturn(func(path string, data []byte) (please.File, error) {
			assert.Equal(t, string(data), path)

			if err, ok := td.ParseErr[path]; ok {
				return nil, err
			}

			file, ok := td.Parse[path]
			if !ok {
				t.Errorf("unexpected call to please parse: %s", path)
			}

			return file, nil
		})

	t.please.EXPECT().NewFile(any).AnyTimes().
		DoAndReturn(func(path string) (please.File, error) {
			file, ok := td.Parse[path]
			if !ok || file != nil {
				t.Errorf("unexpected call to please new file: %s", path)
			}

			return &please.BuildFile{Path: path}, nil
		})

	t.filesystem.EXPECT().ReadAll(any, any).AnyTimes().
		DoAndReturn(func(buf *bytes.Buffer, path string) error {
			file, ok := td.Parse[path]
			if !ok {
				t.Errorf("unexpected call to filesystem read all: %s", path)
			}

			err := td.ParseErr[path]

			if file == nil && err == nil {
				return os.ErrNotExist
			}

			buf.Reset()
			buf.WriteString(path)

			return nil
		})

	t.filesystem.EXPECT().ReadDir(any).AnyTimes().
		DoAndReturn(func(dir string) ([]os.FileInfo, error) {
			var infos []os.FileInfo

			for path, info := range td.Lstat {
				if info != nil && filepath.Dir(path) == dir {
					infos = append(infos, info)
				}
			}

			if infos == nil {
				return nil, os.ErrNotExist
			}

			return infos, nil
		})

	t.filesystem.EXPECT().Walk(any, any).AnyTimes().
		DoAndReturn(func(path string, walkFn filepath.WalkFunc) error {
			for _, path := range td.Walk {
				info := td.Lstat[path]
				if info == nil {
					continue
				}

				if err := walkFn(path, info, nil); err != nil {
					return err
				}
			}

			return nil
		})

	t.filesystem.EXPECT().Lstat(any).AnyTimes().
		DoAndReturn(func(path string) (os.FileInfo, error) {
			info, ok := td.Lstat[path]
			if !ok {
				t.Errorf("unexpected call to filesystem lstat: %s", path)
			}

			if info == nil {
				return nil, os.ErrNotExist
			}

			return info, nil
		})

	t.filesystem.EXPECT().Stat(any).AnyTimes().
		DoAndReturn(func(path string) (os.FileInfo, error) {
			info, ok := td.Stat[path]
			if !ok {
				t.Errorf("unexpected call to filesystem stat: %s", path)
			}

			if info == nil {
				return nil, os.ErrNotExist
			}

			return info, nil
		})

	t.filesystem.EXPECT().Config(any).AnyTimes().
		DoAndReturn(func(path string) *filesystem.Config {
			config, ok := td.Config[path]
			if !ok {
				t.Errorf("unexpected call to filesystem config: %s", path)
			}

			if config == nil {
				config = &filesystem.Config{}
			}

			return config
		})

	t.please.EXPECT().NewRule(any, any).AnyTimes().DoAndReturn(please.NewRule)

	t.please.EXPECT().Write(any).AnyTimes().Do(func(have please.File) {
		path := have.GetPath()

		want, ok := td.Write[path]
		if !ok {
			t.Errorf("unexpected call to please write: %s", path)
		}

		expect.Equal(t, want, have)
	})
}

type GoFormatTestData struct {
	Config    map[string]*filesystem.Config
	ImportDir map[string]*golang.Package
	IsGoroot  map[string]bool
	Lstat     map[string]*FileInfo
	Parse     map[string]*please.BuildFile
	ParseErr  map[string]error
	Stat      map[string]*FileInfo
	Write     map[string]*please.BuildFile
	Readlink  map[string]string
	Walk      []string
	Graph     *please.Graph
}

func (data *GoFormatTestData) Build() {
	for path, pkg := range data.ImportDir {
		data.Stat[path] = &FileInfo{
			FileName:  filepath.Base(path),
			FileMode:  os.FileMode(2147484141),
			FileIsDir: true,
		}

		files := pkg.GoFiles
		files = append(files, pkg.TestGoFiles...)
		files = append(files, pkg.XTestGoFiles...)

		for _, name := range files {
			data.Stat[filepath.Join(path, name)] = &FileInfo{
				FileName: name,
				FileMode: os.FileMode(420),
			}
		}
	}

	for path, _ := range data.Parse {
		data.Stat[path] = &FileInfo{
			FileName: filepath.Base(path),
			FileMode: os.FileMode(420),
		}
	}

	for path, info := range data.Stat {
		if _, ok := data.Lstat[path]; !ok {
			data.Lstat[path] = info
		}
	}

	data.Walk = make([]string, 0, len(data.Lstat))

	for path, _ := range data.Lstat {
		data.Walk = append(data.Walk, path)
	}

	sort.Strings(data.Walk)
}

func (t *ServiceSuite) GoFormatTestData() *GoFormatTestData {
	data := &GoFormatTestData{
		Config: map[string]*filesystem.Config{
			"app":             &filesystem.Config{},
			"app/server":      &filesystem.Config{},
			"app/protos":      &filesystem.Config{},
			"app/protos/mock": &filesystem.Config{},
		},
		IsGoroot: map[string]bool{
			"github.com/golang/mock/gomock":                    false,
			"github.com/golang/protobuf/proto":                 false,
			"github.com/golang/protobuf/proto/ptypes/wrappers": false,
			"github.com/spf13/cobra":                           false,
			"github.com/spf13/pflag":                           false,
			"github.com/stretchr/testify/assert":               false,
			"github.com/stretchr/testify/require":              false,
			"github.com/wollemi_test/app/protos":               false,
			"github.com/wollemi_test/app/protos/mock":          false,
			"github.com/wollemi_test/app/server":               false,
			"google.golang.org/grpc":                           false,
			"google.golang.org/grpc/codes":                     false,
			"google.golang.org/grpc/credentials":               false,
			"google.golang.org/grpc/metadata":                  false,
			"google.golang.org/grpc/status":                    false,
			"testing":                                          true,
			"fmt":                                              true,
			"strings":                                          true,
			"strconv":                                          true,
			"encoding/json":                                    true,
			"database/sql":                                     true,
		},
		ImportDir: map[string]*golang.Package{
			"app/protos": &golang.Package{
				Name: "protos",
				GoFiles: []string{
					"service.pb.go",
					"entities.pb.go",
				},
				Imports: []string{
					"github.com/golang/protobuf/proto",
					"google.golang.org/grpc",
					"google.golang.org/grpc/codes",
					"google.golang.org/grpc/status",
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
				Imports: []string{
					"github.com/golang/mock/gomock",
					"github.com/wollemi_test/app/protos",
					"google.golang.org/grpc",
					"google.golang.org/grpc/metadata",
				},
				GoFileImports: map[string][]string{
					"mock.mg.go": []string{
						"github.com/golang/mock/gomock",
						"github.com/wollemi_test/app/protos",
						"google.golang.org/grpc",
						"google.golang.org/grpc/metadata",
					},
				},
			},
			"app/server": &golang.Package{
				Name: "server",
				GoFiles: []string{
					"server.go",
				},
				Imports: []string{
					"database/sql",
					"encoding/json",
					"github.com/golang/protobuf/proto/ptypes/wrappers",
					"github.com/wollemi_test/app/protos",
					"google.golang.org/grpc",
					"google.golang.org/grpc/credentials",
					"strconv",
					"strings",
				},
				XTestGoFiles: []string{
					"server_test.go",
				},
				XTestImports: []string{
					"github.com/golang/mock/gomock",
					"github.com/golang/protobuf/proto/ptypes/wrappers",
					"github.com/stretchr/testify/assert",
					"github.com/stretchr/testify/require",
					"github.com/wollemi_test/app/protos/mock",
					"testing",
				},
				GoFileImports: map[string][]string{
					"server_test.go": []string{
						"github.com/golang/mock/gomock",
						"github.com/golang/protobuf/proto/ptypes/wrappers",
						"github.com/stretchr/testify/assert",
						"github.com/stretchr/testify/require",
						"github.com/wollemi_test/app/protos/mock",
						"testing",
					},
					"server.go": []string{
						"database/sql",
						"encoding/json",
						"github.com/golang/protobuf/proto/ptypes/wrappers",
						"github.com/wollemi_test/app/protos",
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
				Imports: []string{
					"fmt",
					"github.com/spf13/cobra",
					"github.com/wollemi_test/app/server",
				},
				TestGoFiles: []string{
					"main_test.go",
				},
				TestImports: []string{
					"github.com/golang/mock/gomock",
					"github.com/stretchr/testify/assert",
					"github.com/stretchr/testify/require",
					"testing",
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
						"github.com/wollemi_test/app/server",
					},
				},
			},
		},
		ParseErr: map[string]error{},
		Stat: map[string]*FileInfo{
			"app/server": &FileInfo{
				FileName:  "server",
				FileMode:  os.FileMode(2147484141),
				FileIsDir: true,
			},
			"app/server/server.go": &FileInfo{
				FileName: "server.go",
				FileMode: os.FileMode(420),
			},
			"app/server/server_test.go": &FileInfo{
				FileName: "server_test.go",
				FileMode: os.FileMode(420),
			},
			"app": &FileInfo{
				FileName:  "app",
				FileMode:  os.FileMode(2147484141),
				FileIsDir: true,
			},
			"app/main.go": &FileInfo{
				FileName: "main.go",
				FileMode: os.FileMode(420),
			},
			"app/main_test.go": &FileInfo{
				FileName: "main_test.go",
				FileMode: os.FileMode(420),
			},
			"app/protos": &FileInfo{
				FileName:  "protos",
				FileMode:  os.FileMode(2147484141),
				FileIsDir: true,
			},
			"app/protos/service.pb.go": &FileInfo{
				FileName: "service.pb.go",
				FileMode: os.FileMode(420),
			},
			"app/protos/entities.pb.go": &FileInfo{
				FileName: "entities.pb.go",
				FileMode: os.FileMode(420),
			},
			"app/protos/mock": &FileInfo{
				FileName:  "mock",
				FileMode:  os.FileMode(2147484141),
				FileIsDir: true,
			},
			"app/protos/mock/mock.mg.go": &FileInfo{
				FileName: "mock.mg.go",
				FileMode: os.FileMode(420),
			},
			"third_party/go/google.golang.org/BUILD.plz": &FileInfo{
				FileName: "BUILD.plz",
				FileMode: os.FileMode(420),
			},
			"third_party/go/github.com/stretchr/BUILD.plz": &FileInfo{
				FileName: "BUILD.plz",
				FileMode: os.FileMode(420),
			},
			"third_party/go/github.com/golang/BUILD.plz": &FileInfo{
				FileName: "BUILD.plz",
				FileMode: os.FileMode(420),
			},
			"third_party/go/github.com/spf13/BUILD.plz": &FileInfo{
				FileName: "BUILD.plz",
				FileMode: os.FileMode(420),
			},
			"app/protos/BUILD.plz": &FileInfo{
				FileName: "BUILD.plz",
				FileMode: os.FileMode(420),
			},
			"app/BUILD.plz": &FileInfo{
				FileName: "BUILD.plz",
				FileMode: os.FileMode(420),
			},
			"app/server/BUILD.plz": &FileInfo{
				FileName: "BUILD.plz",
				FileMode: os.FileMode(420),
			},
			"app/protos/mock/BUILD.plz":                                                 nil,
			"third_party/go/github.com/golang/mock/BUILD.plz":                           nil,
			"third_party/go/github.com/golang/mock/gomock/BUILD.plz":                    nil,
			"third_party/go/github.com/golang/protobuf/BUILD.plz":                       nil,
			"third_party/go/github.com/golang/protobuf/proto/BUILD.plz":                 nil,
			"third_party/go/github.com/golang/protobuf/proto/ptypes/BUILD.plz":          nil,
			"third_party/go/github.com/golang/protobuf/proto/ptypes/wrappers/BUILD.plz": nil,
			"third_party/go/github.com/spf13/cobra/BUILD.plz":                           nil,
			"third_party/go/github.com/stretchr/testify/BUILD.plz":                      nil,
			"third_party/go/github.com/stretchr/testify/assert/BUILD.plz":               nil,
			"third_party/go/github.com/stretchr/testify/require/BUILD.plz":              nil,
			"third_party/go/google.golang.org/grpc/BUILD.plz":                           nil,
			"third_party/go/google.golang.org/grpc/codes/BUILD.plz":                     nil,
			"third_party/go/google.golang.org/grpc/credentials/BUILD.plz":               nil,
			"third_party/go/google.golang.org/grpc/metadata/BUILD.plz":                  nil,
			"third_party/go/google.golang.org/grpc/status/BUILD.plz":                    nil,
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
		Parse: map[string]*please.BuildFile{
			"app/BUILD.plz": &please.BuildFile{
				Path: "app/BUILD.plz",
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
				Path: "app/server/BUILD.plz",
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
				Path: "app/protos/BUILD.plz",
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
						please.NewAssignExpr("=", "package", "github.com/wollemi_test/app/protos"),
						please.NewAssignExpr("=", "visibility", []string{"//..."}),
						please.NewAssignExpr("=", "deps", []string{
							":protos",
							"//third_party/go/github.com/golang:mock",
							"//third_party/go/google.golang.org:grpc",
						}),
					}),
				},
			},
			"third_party/go/github.com/spf13/BUILD.plz": &please.BuildFile{
				Path: "third_party/go/github.com/spf13/BUILD.plz",
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
			},
			"third_party/go/github.com/golang/BUILD.plz": &please.BuildFile{
				Path: "third_party/go/github.com/golang/BUILD.plz",
				Stmt: []please.Expr{
					please.NewCallExpr("go_get", []please.Expr{
						please.NewAssignExpr("=", "name", "protobuf"),
						please.NewAssignExpr("=", "get", "github.com/golang/protobuf/..."),
						please.NewAssignExpr("=", "revision", "v1.3.2"),
					}),
					please.NewCallExpr("go_get", []please.Expr{
						please.NewAssignExpr("=", "name", "mock"),
						please.NewAssignExpr("=", "get", "github.com/golang/mock"),
						please.NewAssignExpr("=", "revision", "v1.3.2"),
						please.NewAssignExpr("=", "install", []string{
							"mockgen/model",
							"gomock",
						}),
						please.NewAssignExpr("=", "deps", []string{
							"//third_party/go/golang.org/x:tools",
						}),
					}),
				},
			},
			"third_party/go/github.com/stretchr/BUILD.plz": &please.BuildFile{
				Path: "third_party/go/github.com/stretchr/BUILD.plz",
				Stmt: []please.Expr{
					please.NewCallExpr("go_get", []please.Expr{
						please.NewAssignExpr("=", "name", "testify"),
						please.NewAssignExpr("=", "get", "github.com/stretchr/testify"),
						please.NewAssignExpr("=", "revision", "v1.4.0"),
						please.NewAssignExpr("=", "install", []string{
							"assert",
							"require",
							"vendor/github.com/davecgh/go-spew/spew",
							"vendor/github.com/pmezard/go-difflib/difflib",
						}),
						please.NewAssignExpr("=", "deps", []string{
							"//third_party/go/gopkg.in:yaml.v2",
						}),
					}),
				},
			},
			"third_party/go/google.golang.org/BUILD.plz": &please.BuildFile{
				Path: "third_party/go/google.golang.org/BUILD.plz",
				Stmt: []please.Expr{
					please.NewCallExpr("go_get", []please.Expr{
						please.NewAssignExpr("=", "name", "grpc"),
						please.NewAssignExpr("=", "get", "google.golang.org/grpc/..."),
						please.NewAssignExpr("=", "repo", "github.com/grpc/grpc-go"),
						please.NewAssignExpr("=", "revision", "v1.26.0"),
						please.NewAssignExpr("=", "deps", []string{
							"//third_party/go:genproto_googleapis_rpc_status",
							"//third_party/go/github.com/golang:protobuf",
							"//third_party/go/github.com/google:go-cmp",
							"//third_party/go/golang.org/x:net",
							"//third_party/go/golang.org/x:oauth2",
							"//third_party/go/golang.org/x:sys",
							"//third_party/go/golang.org/x:text",
						}),
					}),
				},
			},
		},
		Write: map[string]*please.BuildFile{},
		Readlink: map[string]string{
			"app/protos/service.pb.go":   "plz-out/gen/app/protos/service.pb.go",
			"app/protos/entities.pb.go":  "plz-out/gen/app/protos/entities.pb.go",
			"app/protos/mock/mock.mg.go": "plz-out/gen/app/protos/mock/mock.mg.go",
		},
	}

	data.Build()

	return data
}

type FileInfo struct {
	FileName    string
	FileSize    int64
	FileMode    os.FileMode
	FileModTime time.Time
	FileIsDir   bool
}

func (this *FileInfo) Name() string {
	return this.FileName
}

func (this *FileInfo) Size() int64 {
	return this.FileSize
}

func (this *FileInfo) Mode() os.FileMode {
	return this.FileMode
}

func (this *FileInfo) ModTime() time.Time {
	return this.FileModTime
}

func (this *FileInfo) IsDir() bool {
	return this.FileIsDir
}

func (this *FileInfo) Sys() interface{} {
	return nil
}
