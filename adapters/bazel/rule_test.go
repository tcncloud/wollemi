package bazel_test

import (
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/adapters/bazel"
	"github.com/tcncloud/wollemi/ports/please"
)

func TestRule_Name(t *testing.T) {
	NewBuilderSuite(t).TestRule_Name()
}

func TestRule_Kind(t *testing.T) {
	NewBuilderSuite(t).TestRule_Kind()
}

func TestRule_SetKind(t *testing.T) {
	NewBuilderSuite(t).TestRule_SetKind()
}

func TestRule_Attr(t *testing.T) {
	NewBuilderSuite(t).TestRule_Attr()
}

func TestRule_SetAttr(t *testing.T) {
	NewBuilderSuite(t).TestRule_SetAttr()
}

func TestRule_DelAttr(t *testing.T) {
	NewBuilderSuite(t).TestRule_DelAttr()
}

// -----------------------------------------------------------------------------

func (t *BuilderSuite) TestRule_Name() {
	type T = BuilderSuite

	t.It("gets call expr attr name", func(t *T) {
		rule := bazel.NewRule(&build.CallExpr{
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "visibility"},
					Op:  "=",
					RHS: &build.ListExpr{
						List: []build.Expr{
							&build.StringExpr{Value: "PUBLIC"},
						},
					},
				},
				&build.AssignExpr{
					LHS: &build.Ident{Name: "name"},
					Op:  "=",
					RHS: &build.StringExpr{Value: "buildtools"},
				},
				&build.AssignExpr{
					LHS: &build.Ident{Name: "external"},
					Op:  "=",
					RHS: &build.Ident{Name: "True"},
				},
			},
		})

		require.Equal(t, "buildtools", rule.Name())
	})
}

func (t *BuilderSuite) TestRule_Kind() {
	type T = BuilderSuite

	t.It("gets call expr ident name", func(t *T) {
		call := bazel.NewRule(&build.CallExpr{
			X: &build.Ident{Name: "foo"},
		})

		require.Equal(t, "foo", call.Kind())
	})
}

func (t *BuilderSuite) TestRule_SetKind() {
	type T = BuilderSuite

	t.It("sets call expr ident name", func(t *T) {
		have := &build.CallExpr{
			X: &build.Ident{Name: "foo"},
		}

		want := &build.CallExpr{
			X: &build.Ident{Name: "bar"},
		}

		bazel.NewRule(have).SetKind("bar")

		require.Equal(t, want, have)
	})
}

func (t *BuilderSuite) TestRule_Attr() {
	type T = BuilderSuite

	t.It("gets attribute value", func(t *T) {
		call := bazel.NewRule(&build.CallExpr{
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "one"},
					RHS: &build.StringExpr{Value: "1"},
				},
				&build.AssignExpr{
					LHS: &build.Ident{Name: "two"},
					RHS: &build.ListExpr{
						List: []build.Expr{
							&build.StringExpr{Value: "2"},
						},
					},
				},
				&build.AssignExpr{
					LHS: &build.Ident{Name: "three"},
					RHS: &build.StringExpr{Value: "3"},
				},
			},
		})

		v1 := call.Attr("one")
		v2 := call.Attr("two")
		v3 := call.Attr("+-1")

		require.Equal(t, please.String("1"), v1)
		require.Equal(t, please.Strings("2"), v2)
		require.Nil(t, v3)
	})
}

func (t *BuilderSuite) TestRule_SetAttr() {
	type T = BuilderSuite

	t.It("overwrites existing attribute value", func(t *T) {
		expr := &build.CallExpr{
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "one"},
					RHS: &build.StringExpr{Value: "1"},
				},
				&build.AssignExpr{
					LHS: &build.Ident{Name: "two"},
					RHS: &build.ListExpr{
						List: []build.Expr{
							&build.StringExpr{Value: "2"},
						},
					},
				},
				&build.AssignExpr{
					LHS: &build.Ident{Name: "three"},
					RHS: &build.StringExpr{Value: "3"},
				},
			},
		}

		call := bazel.NewRule(expr)

		call.SetAttr("three", please.String("333"))

		want := &build.AssignExpr{
			LHS: &build.Ident{Name: "three"},
			RHS: &build.StringExpr{Value: "333"},
		}

		require.Equal(t, want, expr.List[2])
	})

	t.It("creates new attribute", func(t *T) {
		have := &build.CallExpr{}

		bazel.NewRule(have).SetAttr("external", &please.Ident{Name: "True"})

		want := &build.CallExpr{
			List: []build.Expr{
				&build.AssignExpr{
					Op:  "=",
					LHS: &build.Ident{Name: "external"},
					RHS: &build.Ident{Name: "True"},
				},
			},
		}

		require.Equal(t, want, have)
	})

	t.It("keeps marked list exprs", func(t *T) {
		have := &build.CallExpr{
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "deps"},
					RHS: &build.ListExpr{
						List: []build.Expr{
							&build.StringExpr{
								Value: "1",
								Comments: build.Comments{
									Suffix: []build.Comment{{
										Token: "# wollemi:keep",
									}},
								},
							},
							&build.StringExpr{
								Value: "2",
							},
							&build.StringExpr{
								Value: "3",
								Comments: build.Comments{
									Suffix: []build.Comment{{
										Token: "# wollemi:keep",
									}},
								},
							},
						},
					},
				},
			},
		}

		want := &build.CallExpr{
			List: []build.Expr{
				&build.AssignExpr{
					LHS: &build.Ident{Name: "deps"},
					RHS: &build.ListExpr{
						List: []build.Expr{
							&build.StringExpr{
								Value: "1",
								Comments: build.Comments{
									Suffix: []build.Comment{{
										Token: "# wollemi:keep",
									}},
								},
							},
							&build.StringExpr{
								Value: "3",
								Comments: build.Comments{
									Suffix: []build.Comment{{
										Token: "# wollemi:keep",
									}},
								},
							},
							&build.StringExpr{
								Value: "4",
							},
							&build.StringExpr{
								Value: "5",
							},
						},
					},
				},
			},
		}

		bazel.NewRule(have).SetAttr("deps", please.Strings("4", "5"))

		require.Equal(t, want, have)
	})
}

func (t *BuilderSuite) TestRule_DelAttr() {
	type T = BuilderSuite

	t.It("deletes attribute", func(t *T) {
		have := &build.CallExpr{
			List: []build.Expr{
				&build.AssignExpr{
					Op:  "=",
					LHS: &build.Ident{Name: "one"},
					RHS: &build.StringExpr{Value: "1"},
				},
				&build.AssignExpr{
					Op:  "=",
					LHS: &build.Ident{Name: "two"},
					RHS: &build.ListExpr{
						List: []build.Expr{
							&build.StringExpr{Value: "2"},
						},
					},
				},
				&build.AssignExpr{
					Op:  "=",
					LHS: &build.Ident{Name: "three"},
					RHS: &build.StringExpr{Value: "3"},
				},
			},
		}

		// ---------------------------------------------------------------------

		require.NotNil(t, bazel.NewRule(have).DelAttr("two"))

		want := &build.CallExpr{
			List: []build.Expr{
				&build.AssignExpr{
					Op:  "=",
					LHS: &build.Ident{Name: "one"},
					RHS: &build.StringExpr{Value: "1"},
				},
				&build.AssignExpr{
					Op:  "=",
					LHS: &build.Ident{Name: "three"},
					RHS: &build.StringExpr{Value: "3"},
				},
			},
		}

		require.Equal(t, want, have)

		// ---------------------------------------------------------------------

		require.NotNil(t, bazel.NewRule(have).DelAttr("three"))

		want = &build.CallExpr{
			List: []build.Expr{
				&build.AssignExpr{
					Op:  "=",
					LHS: &build.Ident{Name: "one"},
					RHS: &build.StringExpr{Value: "1"},
				},
			},
		}

		require.Equal(t, want, have)
	})
}
