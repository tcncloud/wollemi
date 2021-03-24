package wollemi_test

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/domain/optional"
	"github.com/tcncloud/wollemi/domain/wollemi"
	"github.com/tcncloud/wollemi/ports/golang"
	"github.com/tcncloud/wollemi/testdata/expect"
	"github.com/tcncloud/wollemi/testdata/please"
)

func TestService_GoFormat(t *testing.T) {
	NewServiceSuite(t).TestService_GoFormat()
}

const (
	gosrc = "/go/src"
	gopkg = "github.com/example"
)

func (t *ServiceSuite) TestService_GoFormat() {
	type T = ServiceSuite

	for _, tt := range []struct {
		Title  string
		Config wollemi.Config
		Data   *GoFormatTestData
	}{{ // TEST_CASE -------------------------------------------------------------
		Title: "creates missing go_binary with internal go_test",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app"},
			Parse: t.WithThirdPartyGo(nil),
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:        "main",
					GoFiles:     []string{"main.go"},
					TestGoFiles: []string{"main_test.go"},
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
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
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
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "creates missing go_library with external go_test",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app/server"},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
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
			ImportDir: map[string]*golang.Package{
				"app/server": &golang.Package{
					GoFiles:      []string{"server.go"},
					XTestGoFiles: []string{"server_test.go"},
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
			},
			Write: map[string]*please.BuildFile{
				"app/server/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "server"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
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
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "creates missing go rules across multiple directories",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app/..."},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:        "main",
					GoFiles:     []string{"main.go"},
					TestGoFiles: []string{"main_test.go"},
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
				"app/server": &golang.Package{
					GoFiles:      []string{"server.go"},
					XTestGoFiles: []string{"server_test.go"},
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
			},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
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
							please.NewAssignExpr("=", "visibility", []string{"//app/..."}),
							please.NewAssignExpr("=", "deps", []string{
								":protos",
								"//third_party/go/github.com/golang:mock",
								"//third_party/go/google.golang.org:grpc",
							}),
						}),
					},
				},
			}),
			Write: map[string]*please.BuildFile{
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
							please.NewAssignExpr("=", "visibility", []string{"//app/..."}),
							please.NewAssignExpr("=", "deps", []string{
								":protos",
								"//third_party/go/github.com/golang:mock",
								"//third_party/go/google.golang.org:grpc",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "manages existing go_library rules",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app/server"},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"app/server/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "server"),
							please.NewAssignExpr("=", "srcs", []string{"server.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
					},
				},
			}),
			ImportDir: map[string]*golang.Package{
				"app/server": &golang.Package{
					GoFiles: []string{"server.go"},
					GoFileImports: map[string][]string{
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
			},
			Write: map[string]*please.BuildFile{
				"app/server/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "server"),
							please.NewAssignExpr("=", "srcs", []string{"server.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//app/protos",
								"//third_party/go/github.com/golang:protobuf",
								"//third_party/go/google.golang.org:grpc",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "supports projects with no module name",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: "",
			Paths: []string{"app"},
			Parse: t.WithThirdPartyGo(nil),
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:    "main",
					GoFiles: []string{"main.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{
							"fmt",
							"github.com/spf13/cobra",
							"app/server",
						},
					},
				},
				"app/server": &golang.Package{
					GoFiles: []string{"server.go"},
				},
			},
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//app/server",
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "manages multiple go_library and go_test rules in one build file",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app"},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:         "app",
					GoFiles:      []string{"foo.go", "bar.go"},
					XTestGoFiles: []string{"foo_test.go", "bar_test.go"},
					GoFileImports: map[string][]string{
						"foo.go": []string{
							"database/sql",
							"github.com/spf13/cobra",
						},
						"bar.go": []string{
							"encoding/json",
							"github.com/spf13/pflag",
						},
						"foo_test.go": []string{
							"github.com/example/app",
							"github.com/stretchr/testify/assert",
							"testing",
						},
						"bar_test.go": []string{
							"github.com/example/app",
							"github.com/golang/mock/gomock",
							"testing",
						},
					},
				},
			},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "foo"),
							please.NewAssignExpr("=", "srcs", []string{"foo.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "bar"),
							please.NewAssignExpr("=", "srcs", []string{"bar.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "foo_test"),
							please.NewAssignExpr("=", "srcs", []string{"foo_test.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "external", true),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "bar_test"),
							please.NewAssignExpr("=", "srcs", []string{"bar_test.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "external", true),
						}),
					},
				},
			}),
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "foo"),
							please.NewAssignExpr("=", "srcs", []string{"foo.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "bar"),
							please.NewAssignExpr("=", "srcs", []string{"bar.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:pflag",
							}),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "foo_test"),
							please.NewAssignExpr("=", "srcs", []string{"foo_test.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "external", true),
							please.NewAssignExpr("=", "deps", []string{
								":app",
								"//third_party/go/github.com/stretchr:testify",
							}),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "bar_test"),
							please.NewAssignExpr("=", "srcs", []string{"bar_test.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "external", true),
							please.NewAssignExpr("=", "deps", []string{
								":app",
								"//third_party/go/github.com/golang:mock",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "allows internal go_test to depend on go_library without go import",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app"},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:        "main",
					GoFiles:     []string{"main.go"},
					TestGoFiles: []string{"main_test.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{
							"fmt",
							"github.com/spf13/cobra",
						},
						"main_test.go": []string{
							"github.com/stretchr/testify/assert",
							"testing",
						},
					},
				},
			},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "test"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*_test.go"})),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								":app",
							}),
						}),
					},
				},
			}),
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "test"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*_test.go"})),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								":app",
								"//third_party/go/github.com/stretchr:testify",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "allows unresolved dependencies when configured",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app"},
			Parse: nil,
			Config: map[string]wollemi.Config{
				"app": wollemi.Config{
					AllowUnresolvedDependency: optional.BoolValue(true),
				},
			},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:    "main",
					GoFiles: []string{"main.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{
							"github.com/spf13/cobra",
							"github.com/spf13/pflag",
						},
					},
				},
			},
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "uses go_library import_path when resolving dependencies",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app"},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"third_party/go/github.com/spf13/cobra/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "cobra"),
							please.NewAssignExpr("=", "import_path", "github.com/spf13/cobra"),
						}),
					},
				},
			}),
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:    "main",
					GoFiles: []string{"main.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{
							"github.com/spf13/cobra",
							"github.com/spf13/pflag",
						},
					},
				},
			},
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:pflag",
								"//third_party/go/github.com/spf13/cobra",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "creates go rules with explicit sources when configured",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app/..."},
			Parse: t.WithThirdPartyGo(nil),
			Config: map[string]wollemi.Config{
				"app": wollemi.Config{
					ExplicitSources: optional.BoolValue(true),
				},
				"app/server": wollemi.Config{
					ExplicitSources: optional.BoolValue(true),
				},
			},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:    "main",
					GoFiles: []string{"app_x.go", "app_y.go"},
					GoFileImports: map[string][]string{
						"app_x.go": []string{"github.com/example/app/server"},
						"app_y.go": []string{"github.com/spf13/cobra"},
					},
				},
				"app/server": &golang.Package{
					GoFiles:      []string{"server_x.go", "server_y.go"},
					XTestGoFiles: []string{"server_x_test.go", "server_y_test.go"},
					GoFileImports: map[string][]string{
						"server_x.go":      []string{"google.golang.org/grpc"},
						"server_y.go":      []string{},
						"server_x_test.go": []string{"github.com/stretchr/testify/assert"},
						"server_y_test.go": []string{"github.com/example/app/server"},
					},
				},
			},
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"app_x.go", "app_y.go"}),
							please.NewAssignExpr("=", "visibility", []string{"//app/..."}),
							please.NewAssignExpr("=", "deps", []string{
								"//app/server",
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
				"app/server/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "server"),
							please.NewAssignExpr("=", "srcs", []string{"server_x.go", "server_y.go"}),
							please.NewAssignExpr("=", "visibility", []string{"//app/..."}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/google.golang.org:grpc",
							}),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "test"),
							please.NewAssignExpr("=", "srcs", []string{"server_x_test.go", "server_y_test.go"}),
							please.NewAssignExpr("=", "external", true),
							please.NewAssignExpr("=", "deps", []string{
								":server",
								"//third_party/go/github.com/stretchr:testify",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "manages go rules with explicit sources when configured",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app/..."},
			Config: map[string]wollemi.Config{
				"app": wollemi.Config{
					ExplicitSources: optional.BoolValue(true),
				},
				"app/server": wollemi.Config{
					ExplicitSources: optional.BoolValue(true),
				},
			},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:    "main",
					GoFiles: []string{"app_x.go", "app_y.go"},
					GoFileImports: map[string][]string{
						"app_x.go": []string{"github.com/example/app/server"},
						"app_y.go": []string{"github.com/spf13/cobra"},
					},
				},
				"app/server": &golang.Package{
					GoFiles:     []string{"server_x.go", "server_y.go"},
					TestGoFiles: []string{"server_x_test.go", "server_y_test.go"},
					GoFileImports: map[string][]string{
						"server_x.go":      []string{"google.golang.org/grpc"},
						"server_y.go":      []string{},
						"server_x_test.go": []string{"github.com/stretchr/testify/assert"},
						"server_y_test.go": []string{"testing"},
					},
				},
			},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"app_x.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
					},
				},
				"app/server/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "server"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"}, "*_test.go")),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "test"),
							please.NewAssignExpr("=", "srcs", please.NewGlob([]string{"*.go"})),
							please.NewAssignExpr("=", "deps", []string{":server"}),
						}),
					},
				},
			}),
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"app_x.go", "app_y.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//app/server",
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
				"app/server/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_library", []please.Expr{
							please.NewAssignExpr("=", "name", "server"),
							please.NewAssignExpr("=", "srcs", []string{"server_x.go", "server_y.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/google.golang.org:grpc",
							}),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "test"),
							please.NewAssignExpr("=", "srcs", []string{"server_x_test.go", "server_y_test.go"}),
							please.NewAssignExpr("=", "deps", []string{
								":server",
								"//third_party/go/github.com/stretchr:testify",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "deletes internal go_test rules that lack source test files",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app"},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:    "main",
					GoFiles: []string{"main.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{"github.com/spf13/cobra"},
					},
				},
			},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
						please.NewCallExpr("go_test", []please.Expr{
							please.NewAssignExpr("=", "name", "test"),
							please.NewAssignExpr("=", "srcs", []string{"main.go", "main_test.go"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
			}),
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "does not create missing go rules when configured",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app/..."},
			Config: map[string]wollemi.Config{
				"app/cmd": wollemi.Config{
					Gofmt: wollemi.Gofmt{
						Create: []string{},
					},
				},
			},
			ImportDir: map[string]*golang.Package{
				"app/cmd": &golang.Package{
					Name:    "main",
					GoFiles: []string{"main.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{"github.com/spf13/cobra"},
					},
				},
			},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"cmd/main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
			}),
			Write: map[string]*please.BuildFile{
				"app/cmd/BUILD.plz": &please.BuildFile{},
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"cmd/main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "overrides filesystem package config gofmt create using ctl config",
		Config: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Create: []string{},
			},
		},
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app/..."},
			ImportDir: map[string]*golang.Package{
				"app/cmd": &golang.Package{
					Name:    "main",
					GoFiles: []string{"main.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{"github.com/spf13/cobra"},
					},
				},
			},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"cmd/main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
			}),
			Write: map[string]*please.BuildFile{
				"app/cmd/BUILD.plz": &please.BuildFile{},
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"cmd/main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "manages custom rule kinds when configured through ctl",
		Config: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Manage: []string{"go_custom_binary", "go_custom_test"},
			},
		},
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app/..."},
			Config: map[string]wollemi.Config{
				"app": wollemi.Config{
					Gofmt: wollemi.Gofmt{
						Manage: []string{}, // package disabled; ctl should override
					},
				},
			},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:        "main",
					GoFiles:     []string{"main.go"},
					TestGoFiles: []string{"main_test.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{"github.com/spf13/cobra"},
						"main_test.go": []string{
							"github.com/stretchr/testify",
							"testing",
						},
					},
				},
			},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_custom_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
						please.NewCallExpr("go_custom_test", []please.Expr{
							please.NewAssignExpr("=", "name", "test"),
							please.NewAssignExpr("=", "srcs", []string{"main_test.go"}),
						}),
					},
				},
			}),
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_custom_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
						please.NewCallExpr("go_custom_test", []please.Expr{
							please.NewAssignExpr("=", "name", "test"),
							please.NewAssignExpr("=", "srcs", []string{"main_test.go"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/stretchr:testify",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "maps created rule kinds when configured",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: gopkg,
			Paths: []string{"app/..."},
			Config: map[string]wollemi.Config{
				"app": wollemi.Config{
					ExplicitSources: optional.BoolValue(true),
					Gofmt: wollemi.Gofmt{
						Mapped: map[string]string{
							"go_binary": "go_custom_binary",
							"go_test":   "go_custom_test",
						},
					},
				},
			},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:        "main",
					GoFiles:     []string{"main.go"},
					TestGoFiles: []string{"main_test.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{"github.com/spf13/cobra"},
						"main_test.go": []string{
							"github.com/stretchr/testify",
							"testing",
						},
					},
				},
			},
			Parse: t.WithThirdPartyGo(nil),
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_custom_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"//app/..."}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
						please.NewCallExpr("go_custom_test", []please.Expr{
							please.NewAssignExpr("=", "name", "test"),
							please.NewAssignExpr("=", "srcs", []string{"main.go", "main_test.go"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/github.com/spf13:cobra",
								"//third_party/go/github.com/stretchr:testify",
							}),
						}),
					},
				},
			},
		},
	}, { // TEST_CASE ------------------------------------------------------------
		Title: "resolves internal generated grpc libraries",
		Data: &GoFormatTestData{
			Gosrc: gosrc,
			Gopkg: "",
			Paths: []string{"app"},
			ImportDir: map[string]*golang.Package{
				"app": &golang.Package{
					Name:    "main",
					GoFiles: []string{"main.go"},
					GoFileImports: map[string][]string{
						"main.go": []string{
							"database/sql",
							"fmt",
							"github.com/spf13/cobra",
							"protos/bar",
							"protos/foo",
						},
					},
				},
			},
			Parse: t.WithThirdPartyGo(map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
						}),
					},
				},
				"protos/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("grpc_library", []please.Expr{
							please.NewAssignExpr("=", "name", "foo"),
							please.NewAssignExpr("=", "srcs", []string{"foo.proto"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/google.golang.org:grpc",
							}),
						}),
						please.NewCallExpr("grpc_library", []please.Expr{
							please.NewAssignExpr("=", "name", "bar"),
							please.NewAssignExpr("=", "srcs", []string{"bar.proto"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//third_party/go/google.golang.org:grpc",
							}),
						}),
					},
				},
			}),
			Write: map[string]*please.BuildFile{
				"app/BUILD.plz": &please.BuildFile{
					Stmt: []please.Expr{
						please.NewCallExpr("go_binary", []please.Expr{
							please.NewAssignExpr("=", "name", "app"),
							please.NewAssignExpr("=", "srcs", []string{"main.go"}),
							please.NewAssignExpr("=", "visibility", []string{"PUBLIC"}),
							please.NewAssignExpr("=", "deps", []string{
								"//protos:bar",
								"//protos:foo",
								"//third_party/go/github.com/spf13:cobra",
							}),
						}),
					},
				},
			},
		},
	}} {
		focus := ""

		if !regexp.MustCompile(focus).MatchString(tt.Title) {
			continue
		}

		t.Run(tt.Title, func(t *T) {
			write := make(chan please.File, 1000)

			t.MockGoFormat(tt.Data, write)

			wollemi := t.New(tt.Data.Gosrc, tt.Data.Gopkg)

			require.NoError(t, wollemi.GoFormat(tt.Config, tt.Data.Paths))
			close(write)

			for have := range write {
				path := have.GetPath()
				want := tt.Data.Write[path]

				expect.Equal(t, want, have)
				delete(tt.Data.Write, path)
			}

			for _, want := range tt.Data.Write {
				expect.Equal(t, want, (*please.BuildFile)(nil))
			}
		})
	}
}

func (t *ServiceSuite) MockGoFormat(data *GoFormatTestData, write chan please.File) {
	data.Prepare()

	t.golang.EXPECT().ImportDir(any, any).AnyTimes().
		DoAndReturn(func(path string, names []string) (*golang.Package, error) {
			gopkg, ok := data.ImportDir[path]
			if !ok {
				t.Errorf("unexpected call to golang import dir: %s", path)
			}

			return gopkg, nil
		})

	t.golang.EXPECT().IsGoroot(any).AnyTimes().
		DoAndReturn(func(path string) bool { return data.IsGoroot[path] })

	t.please.EXPECT().Parse(any, any).AnyTimes().
		DoAndReturn(func(path string, buf []byte) (please.File, error) {
			assert.Equal(t, string(buf), path)

			if err, ok := data.ParseErr[path]; ok {
				return nil, err
			}

			file, ok := data.Parse[path]
			if !ok {
				t.Errorf("unexpected call to please parse: %s", path)
			}

			return file, nil
		})

	t.please.EXPECT().NewFile(any).AnyTimes().
		DoAndReturn(func(path string) (please.File, error) {
			file, _ := data.Parse[path]
			if file != nil {
				t.Errorf("unexpected call to please new file: %s", path)
			}

			return &please.BuildFile{Path: path}, nil
		})

	t.filesystem.EXPECT().ReadAll(any, any).AnyTimes().
		DoAndReturn(func(buf *bytes.Buffer, path string) error {
			file, ok := data.Parse[path]
			if !ok {
				t.Errorf("unexpected call to filesystem read all: %s", path)
			}

			err := data.ParseErr[path]

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

			for path, info := range data.Lstat {
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

	t.filesystem.EXPECT().Lstat(any).AnyTimes().
		DoAndReturn(func(path string) (os.FileInfo, error) {
			info, ok := data.Lstat[path]
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
			info, ok := data.Stat[path]
			if !ok {
				return nil, os.ErrNotExist
			}

			return info, nil
		})

	t.filesystem.EXPECT().Config(any).AnyTimes().
		DoAndReturn(func(path string) wollemi.Config {
			return data.Config[path]
		})

	t.please.EXPECT().NewRule(any, any).AnyTimes().DoAndReturn(please.NewRule)

	t.please.EXPECT().Write(any).AnyTimes().
		Do(func(have please.File) { write <- have })
}

type GoFormatTestData struct {
	Gosrc     string
	Gopkg     string
	Paths     []string
	Config    map[string]wollemi.Config
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

// getFileImports gets combined list of imports from the provided files.
func getFileImports(files []string, fileImports map[string][]string) []string {
	have := make(map[string]bool)

	var out []string

	for _, name := range files {
		for _, path := range fileImports[name] {
			if have[path] {
				continue
			}

			out = append(out, path)
			have[path] = true
		}
	}

	sort.Strings(out)

	return out
}

func (d *GoFormatTestData) Prepare() {
	if d.Config == nil {
		d.Config = make(map[string]wollemi.Config)
	}

	if d.ImportDir == nil {
		d.ImportDir = make(map[string]*golang.Package)
	}

	if d.IsGoroot == nil {
		d.IsGoroot = map[string]bool{
			"testing":       true,
			"fmt":           true,
			"strings":       true,
			"strconv":       true,
			"encoding/json": true,
			"database/sql":  true,
		}
	}

	if d.Lstat == nil {
		d.Lstat = make(map[string]*FileInfo)
	}

	if d.Stat == nil {
		d.Stat = make(map[string]*FileInfo)
	}

	if d.Parse == nil {
		d.Parse = make(map[string]*please.BuildFile)
	}

	if d.ParseErr == nil {
		d.ParseErr = make(map[string]error)
	}

	if d.Write == nil {
		d.Write = make(map[string]*please.BuildFile)
	}

	if d.Readlink == nil {
		d.Readlink = make(map[string]string)
	}

	for path, pkg := range d.ImportDir {
		if pkg.Name == "" {
			pkg.Name = filepath.Base(path)
		}

		// Synchronize package import lists from go file imports.
		pkg.Imports = getFileImports(pkg.GoFiles, pkg.GoFileImports)
		pkg.TestImports = getFileImports(pkg.TestGoFiles, pkg.GoFileImports)
		pkg.XTestImports = getFileImports(pkg.XTestGoFiles, pkg.GoFileImports)

		// Setup stat info go package directory.
		d.Stat[path] = &FileInfo{
			FileMode:  os.FileMode(2147484141),
			FileName:  filepath.Base(path),
			FileIsDir: true,
		}

		// Setup stat info for each defined go file.
		for _, files := range [][]string{
			pkg.XTestGoFiles,
			pkg.TestGoFiles,
			pkg.GoFiles,
		} {
			for _, name := range files {
				d.Stat[filepath.Join(path, name)] = &FileInfo{
					FileMode: os.FileMode(420),
					FileName: name,
				}
			}
		}
	}

	// Setup stat info for each expected buildfile parse.
	for path, _ := range d.Parse {
		d.Stat[path] = &FileInfo{
			FileName: filepath.Base(path),
			FileMode: os.FileMode(420),
		}
	}

	for _, buildfiles := range []map[string]*please.BuildFile{d.Parse, d.Write} {
		for path, buildfile := range buildfiles {
			if buildfile != nil && buildfile.Path == "" {
				buildfile.Path = path
			}
		}
	}

	for path, _ := range d.Stat {
		dir := filepath.Dir(path)

		for ; dir != "." && dir != "/"; dir = filepath.Dir(dir) {
			if _, ok := d.Config[dir]; !ok {
				d.Config[dir] = wollemi.Config{}
			}

			// Define stat info for undefined parent directory.
			if _, ok := d.Stat[dir]; !ok {
				d.Stat[dir] = &FileInfo{
					FileName:  filepath.Base(dir),
					FileMode:  os.FileMode(2147484141),
					FileIsDir: true,
				}
			}
		}
	}

	for path, info := range d.Stat {
		if _, ok := d.Lstat[path]; !ok {
			d.Lstat[path] = info
		}
	}

	d.Walk = make([]string, 0, len(d.Lstat))

	for _, walkRoot := range d.Paths {
		walkRoot = strings.TrimSuffix(walkRoot, "/...")

		for path, _ := range d.Lstat {
			if strings.HasPrefix(path, walkRoot) {
				d.Walk = append(d.Walk, path)
			}
		}
	}

	sort.Strings(d.Walk)
}

// WithThirdPartyGo merges the provided extra build files into a default set of
// third party go build files.
func (t *ServiceSuite) WithThirdPartyGo(extra map[string]*please.BuildFile) map[string]*please.BuildFile {
	files := map[string]*please.BuildFile{
		"third_party/go/github.com/spf13/BUILD.plz": &please.BuildFile{
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
	}

	for k, v := range extra {
		files[k] = v
	}

	return files
}

type FileInfo struct {
	FileName    string      `json:"file_name,omitempty"`
	FileSize    int64       `json:"file_size,omitempty"`
	FileMode    os.FileMode `json:"file_mode,omitempty"`
	FileModTime time.Time   `json:"file_mod_time,omitempty"`
	FileIsDir   bool        `json:"file_is_dir,omitempty"`
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
