package bazel_test

import (
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/adapters/bazel"
	"github.com/tcncloud/wollemi/ports/please"
)

func TestFile_GetPath(t *testing.T) {
	NewBuilderSuite(t).TestFile_GetPath()
}

func TestFile_GetRule(t *testing.T) {
	NewBuilderSuite(t).TestFile_GetRule()
}

func TestFile_GetRules(t *testing.T) {
	NewBuilderSuite(t).TestFile_GetRules()
}

func TestFile_SetRule(t *testing.T) {
	NewBuilderSuite(t).TestFile_SetRule()
}

func TestFile_DelRule(t *testing.T) {
	NewBuilderSuite(t).TestFile_DelRule()
}

func (t *BuilderSuite) TestFile_GetPath() {
	type T = BuilderSuite

	t.It("gets parsed file path", func(t *T) {
		path := "foo/bar/BUILD.plz"
		file, err := t.builder.Parse(path, buildtools)
		require.NoError(t, err)
		require.Equal(t, path, file.GetPath())
	})
}

func (t *BuilderSuite) TestFile_GetRule() {
	type T = BuilderSuite

	t.It("gets rule by name", func(t *T) {
		file, err := t.builder.Parse("", buildtools)
		require.NoError(t, err)

		have := file.GetRule("buildtools")
		want := bazel.NewRule(t.BuildtoolsRule(1).(*build.CallExpr))

		require.Equal(t, want, have)
	})
}

func (t *BuilderSuite) TestFile_GetRules() {
	type T = BuilderSuite

	t.It("iterates over all rules", func(t *T) {
		var i int

		file, err := t.builder.Parse("", buildtools)
		require.NoError(t, err)

		file.GetRules(func(have please.Rule) {
			want := bazel.NewRule(t.BuildtoolsRule(i).(*build.CallExpr))

			require.Equalf(t, want, have, "i: %d", i)
			i++
		})

		require.Equal(t, 2, i)
	})
}

func (t *BuilderSuite) TestFile_SetRule() {
	type T = BuilderSuite

	t.It("overwrites existing named rule", func(t *T) {
		file, err := t.builder.Parse("BUILD.plz", buildtools)
		require.NoError(t, err)
		require.IsType(t, &bazel.File{}, file)

		have := (file.(*bazel.File)).Unwrap()

		rule := bazel.NewRule(&build.CallExpr{
			X: &build.Ident{Name: "go_mock"},
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "name"},
					Op:  "=",
					RHS: &build.StringExpr{
						Value: "buildtools",
					},
				},
			},
		})

		file.SetRule(rule)

		want := t.Buildtools()
		want.Stmt[1] = &build.CallExpr{
			X: &build.Ident{Name: "go_mock"},
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "name"},
					Op:  "=",
					RHS: &build.StringExpr{
						Value: "buildtools",
					},
				},
			},
		}

		require.Equal(t, want, have)
	})

	t.It("appends rule when no existing named rule", func(t *T) {
		file, err := t.builder.Parse("BUILD.plz", buildtools)
		require.NoError(t, err)
		require.IsType(t, &bazel.File{}, file)

		have := (file.(*bazel.File)).Unwrap()

		rule := bazel.NewRule(&build.CallExpr{
			X: &build.Ident{Name: "go_mock"},
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "name"},
					Op:  "=",
					RHS: &build.StringExpr{
						Value: "yourself",
					},
				},
			},
		})

		file.SetRule(rule)

		want := t.Buildtools()
		want.Stmt = append(want.Stmt, &build.CallExpr{
			X: &build.Ident{Name: "go_mock"},
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "name"},
					Op:  "=",
					RHS: &build.StringExpr{
						Value: "yourself",
					},
				},
			},
		})

		require.Equal(t, want, have)
	})
}

func (t *BuilderSuite) TestFile_DelRule() {
	type T = BuilderSuite

	t.It("deletes existing named rule", func(t *T) {
		file, err := t.builder.Parse("BUILD.plz", buildtools)
		require.NoError(t, err)
		require.IsType(t, &bazel.File{}, file)

		have := (file.(*bazel.File)).Unwrap()

		file.DelRule("buildtools")

		want := t.Buildtools()
		want.Stmt = want.Stmt[:1]

		require.Equal(t, want, have)
	})
}
