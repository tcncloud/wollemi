package bazel_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/adapters/bazel"
)

func TestBuilder_Parse(t *testing.T) {
	NewBuilderSuite(t).TestBuilder_Parse()
}

func TestBuilder_NewFile(t *testing.T) {
	NewBuilderSuite(t).TestBuilder_NewFile()
}

func TestBuilder_NewRule(t *testing.T) {
	NewBuilderSuite(t).TestBuilder_NewRule()
}

func TestBuilder_Write(t *testing.T) {
	NewBuilderSuite(t).TestBuilder_Write()
}

func (t *BuilderSuite) TestBuilder_Parse() {
	type T = BuilderSuite

	t.It("does not error when parsing f-strings", func(t *T) {
		_, err := t.builder.Parse("BUILD.plz", fstring)
		require.NoError(t, err)
	})

	t.It("parses build file", func(t *T) {
		file, err := t.builder.Parse("BUILD.plz", buildtools)
		require.NoError(t, err)
		require.IsType(t, &bazel.File{}, file)

		have := (file.(*bazel.File)).Unwrap()
		want := t.Buildtools()

		require.Equal(t, want, have)
	})
}

func (t *BuilderSuite) TestBuilder_NewFile() {
	type T = BuilderSuite

	t.It("creates new empty build file", func(t *T) {
		file := t.builder.NewFile("foo/bar/BUILD.plz")
		require.IsType(t, &bazel.File{}, file)

		have := (file.(*bazel.File)).Unwrap()
		want := &build.File{
			Path: "foo/bar/BUILD.plz",
			Type: build.TypeBuild,
		}

		require.Equal(t, want, have)
	})
}

func (t *BuilderSuite) TestBuilder_NewRule() {
	type T = BuilderSuite

	t.It("creates new empty build rule", func(t *T) {
		have := t.builder.NewRule("go_mock", "yourself")
		want := bazel.NewRule(&build.CallExpr{
			X: &build.Ident{Name: "go_mock"},
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "name"},
					Op:  "=",
					RHS: &build.StringExpr{Value: "yourself"},
				},
			},
		})

		require.Equal(t, want, have)
	})
}

func (t *BuilderSuite) TestBuilder_Write() {
	type T = BuilderSuite

	t.It("deletes build file when it lacks named rules", func(t *T) {
		file := bazel.NewFile(nil, &build.File{
			Path: "wollemi/ports/BUILD.plz",
			Stmt: []build.Expr{
				&build.CallExpr{
					X: &build.Ident{Name: "subinclude"},
					List: []build.Expr{
						&build.StringExpr{Value: "//build_defs:go"},
					},
				},
				&build.CallExpr{
					X: &build.Ident{Name: "package"},
					List: []build.Expr{
						&build.AssignExpr{
							Op:  "=",
							LHS: &build.Ident{Name: "default_visibility"},
							RHS: &build.ListExpr{
								List: []build.Expr{
									&build.StringExpr{Value: "//..."},
								},
							},
						},
					},
				},
			},
		})

		// Empty directories containing the build file should be removed.
		gomock.InOrder(
			t.filesystem.EXPECT().Remove("wollemi/ports/BUILD.plz"),
			t.filesystem.EXPECT().ReadDir("wollemi/ports"),
			t.filesystem.EXPECT().Remove("wollemi/ports"),
			t.filesystem.EXPECT().ReadDir("wollemi").
				Return(make([]os.FileInfo, 1), nil),
		)

		require.NoError(t, t.builder.Write(file))
	})

	t.It("writes formatted build file", func(t *T) {
		file := bazel.NewFile(nil, &build.File{
			Path: "wollemi/ports/BUILD.plz",
			Stmt: []build.Expr{
				&build.CallExpr{
					X: &build.Ident{Name: "subinclude"},
					List: []build.Expr{
						&build.StringExpr{Value: "//build_defs:go"},
					},
				},
				&build.CallExpr{
					X: &build.Ident{Name: "package"},
					List: []build.Expr{
						&build.AssignExpr{
							Op:  "=",
							LHS: &build.Ident{Name: "default_visibility"},
							RHS: &build.ListExpr{
								List: []build.Expr{
									&build.StringExpr{Value: "//..."},
								},
							},
						},
					},
				},
				&build.CallExpr{
					X: &build.Ident{Name: "go_mock"},
					List: []build.Expr{
						&build.AssignExpr{
							Op:  "=",
							LHS: &build.Ident{Name: "name"},
							RHS: &build.StringExpr{Value: "mock_please"},
						},
						&build.AssignExpr{
							Op:  "=",
							LHS: &build.Ident{Name: "ginkgo"},
							RHS: &build.Ident{Name: "False"},
						},
						&build.AssignExpr{
							Op:  "=",
							LHS: &build.Ident{Name: "interfaces"},
							RHS: &build.ListExpr{
								List: []build.Expr{
									&build.StringExpr{Value: "Filesystem"},
								},
							},
						},
						&build.AssignExpr{
							Op:  "=",
							LHS: &build.Ident{Name: "package"},
							RHS: &build.StringExpr{Value: "github.com/tcncloud/wollemi/ports/please"},
						},
						&build.AssignExpr{
							Op:  "=",
							LHS: &build.Ident{Name: "deps"},
							RHS: &build.ListExpr{
								List: []build.Expr{
									&build.StringExpr{Value: "//ports/please"},
								},
							},
						},
					},
				},
			},
		})

		var want bytes.Buffer

		want.WriteString("subinclude(\"//build_defs:go\")\n")
		want.WriteString("package(default_visibility = [\"//...\"])\n")
		want.WriteString("go_mock(\n")
		want.WriteString("    name = \"mock_please\",\n")
		want.WriteString("    ginkgo = False,\n")
		want.WriteString("    interfaces = [\"Filesystem\"],\n")
		want.WriteString("    package = \"github.com/tcncloud/wollemi/ports/please\",\n")
		want.WriteString("    deps = [\"//ports/please\"],\n")
		want.WriteString(")\n")

		t.filesystem.EXPECT().WriteFile(any, any, any).
			Do(func(path string, data []byte, mode os.FileMode) {
				require.Equal(t, "wollemi/ports/BUILD.plz", path)
				require.Equal(t, want.String(), string(data))
				require.Equal(t, os.FileMode(0644), mode)
			})

		require.NoError(t, t.builder.Write(file))
	})
}

var buildtools = []byte(`
package(default_visibility = ['PUBLIC'])

go_get(
    name = 'buildtools',
    get = 'github.com/bazelbuild/buildtools',
    repo = 'github.com/bazelbuild/buildtools',
    revision = '3.0.0',
    install = [
        'build',
        'tables',
    ],
)
`)

var fstring = []byte(`
package(default_visibility = ['PUBLIC'])

subinclude('//build_defs:go')
revision = 'v0.44.3'

get = 'cloud.google.com/go/'

remote_file(
    name ='_google-cloud-go#download',
    url = f'https://github.com/googleapis/google-cloud-go/archive/{revision}.tar.gz',
    out = 'google-cloud-go.tar.gz',
)
`)
