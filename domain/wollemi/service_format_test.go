package wollemi_test

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
}

func (t *ServiceSuite) MockGoFormat(td *GoFormatTestData) {
	t.golang.EXPECT().ImportDir(any).AnyTimes().
		DoAndReturn(func(path string) (*golang.Package, error) {
			gopkg, ok := td.ImportDir[path]
			if !ok {
				t.Errorf("unexpected call to golang import dir: %s", path)
			}

			return gopkg, nil
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
	Lstat     map[string]*FileInfo
	Parse     map[string]*please.BuildFile
	ParseErr  map[string]error
	Stat      map[string]*FileInfo
	Write     map[string]*please.BuildFile
	Readlink  map[string]string
	Walk      []string
	Graph     *please.Graph
}

func (t *ServiceSuite) GoFormatTestData() *GoFormatTestData {
	data := &GoFormatTestData{
		Config: map[string]*filesystem.Config{
			"app":             &filesystem.Config{},
			"app/server":      &filesystem.Config{},
			"app/protos":      &filesystem.Config{},
			"app/protos/mock": &filesystem.Config{},
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
			},
			"app/server": &golang.Package{
				Name: "server",
				GoFiles: []string{
					"server.go",
				},
				Imports: []string{
					"github.com/wollemi_test/app/protos",
					"github.com/golang/protobuf/proto/ptypes/wrappers",
					"google.golang.org/grpc",
					"google.golang.org/grpc/credentials",
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
				},
			},
			"app": &golang.Package{
				Name: "main",
				GoFiles: []string{
					"main.go",
				},
				Imports: []string{
					"github.com/wollemi_test/app/server",
					"github.com/spf13/cobra",
				},
				TestGoFiles: []string{
					"main_test.go",
				},
				TestImports: []string{
					"github.com/golang/mock/gomock",
					"github.com/stretchr/testify/assert",
					"github.com/stretchr/testify/require",
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
