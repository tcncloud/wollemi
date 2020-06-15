package bazel_test

import (
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/golang/mock/gomock"

	"github.com/tcncloud/wollemi/adapters/bazel"
	"github.com/tcncloud/wollemi/ports/please/mock"
	"github.com/tcncloud/wollemi/testdata/mem"
)

var any = gomock.Any()

func NewBuilderSuite(t *testing.T) *BuilderSuite {
	return &BuilderSuite{T: t}
}

type BuilderSuite struct {
	*testing.T
	builder    *bazel.Builder
	logger     *mem.Logger
	ctl        *mock_please.MockCtl
	filesystem *mock_please.MockFilesystem
}

func (suite *BuilderSuite) It(name string, yield func(*BuilderSuite)) {
	suite.Helper()
	suite.Run(name, yield)
}

func (suite *BuilderSuite) Run(name string, yield func(*BuilderSuite)) {
	suite.Helper()
	suite.T.Run(name, func(t *testing.T) {
		suite := NewBuilderSuite(t)
		suite.Helper()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		suite.logger = mem.NewLogger()
		suite.filesystem = mock_please.NewMockFilesystem(ctrl)
		suite.ctl = mock_please.NewMockCtl(ctrl)

		suite.builder = bazel.NewBuilder(suite.logger, suite.ctl, suite.filesystem)

		yield(suite)
	})
}

func (suite *BuilderSuite) DefaultMocks() {
	suite.filesystem.EXPECT().WriteFile(any, any, any).AnyTimes()
	suite.filesystem.EXPECT().Remove(any).AnyTimes()
}

func (suite *BuilderSuite) Buildtools() *build.File {
	return &build.File{
		Path: "BUILD.plz",
		Type: build.TypeBuild,
		Stmt: []build.Expr{
			&build.CallExpr{
				X: &build.Ident{
					NamePos: build.Position{
						Line:     2,
						LineRune: 1,
						Byte:     1,
					},
					Name: "package",
				},
				ListStart: build.Position{
					Line:     2,
					LineRune: 8,
					Byte:     8,
				},
				List: []build.Expr{
					&build.AssignExpr{
						LHS: &build.Ident{
							NamePos: build.Position{
								Line:     2,
								LineRune: 9,
								Byte:     9,
							},
							Name: "default_visibility",
						},
						OpPos: build.Position{
							Line:     2,
							LineRune: 28,
							Byte:     28,
						},
						Op: "=",
						RHS: &build.ListExpr{
							Start: build.Position{
								Line:     2,
								LineRune: 30,
								Byte:     30,
							},
							List: []build.Expr{
								&build.StringExpr{
									Start: build.Position{
										Line:     2,
										LineRune: 31,
										Byte:     31,
									},
									Value: "PUBLIC",
									End: build.Position{
										Line:     2,
										LineRune: 39,
										Byte:     39,
									},
									Token: "'PUBLIC'",
								},
							},
							End: build.End{
								Pos: build.Position{
									Line:     2,
									LineRune: 39,
									Byte:     39,
								},
							},
						},
					},
				},
				End: build.End{
					Pos: build.Position{
						Line:     2,
						LineRune: 40,
						Byte:     40,
					},
				},
			},
			&build.CallExpr{
				X: &build.Ident{
					NamePos: build.Position{
						Line:     4,
						LineRune: 1,
						Byte:     43,
					},
					Name: "go_get",
				},
				ListStart: build.Position{
					Line:     4,
					LineRune: 7,
					Byte:     49,
				},
				List: []build.Expr{
					&build.AssignExpr{
						LHS: &build.Ident{
							NamePos: build.Position{
								Line:     5,
								LineRune: 5,
								Byte:     55,
							},
							Name: "name",
						},
						OpPos: build.Position{
							Line:     5,
							LineRune: 10,
							Byte:     60,
						},
						Op: "=",
						RHS: &build.StringExpr{
							Start: build.Position{
								Line:     5,
								LineRune: 12,
								Byte:     62,
							},
							Value: "buildtools",
							End: build.Position{
								Line:     5,
								LineRune: 24,
								Byte:     74,
							},
							Token: "'buildtools'",
						},
					},
					&build.AssignExpr{
						LHS: &build.Ident{
							NamePos: build.Position{
								Line:     6,
								LineRune: 5,
								Byte:     80,
							},
							Name: "get",
						},
						OpPos: build.Position{
							Line:     6,
							LineRune: 9,
							Byte:     84,
						},
						Op: "=",
						RHS: &build.StringExpr{
							Start: build.Position{
								Line:     6,
								LineRune: 11,
								Byte:     86,
							},
							Value: "github.com/bazelbuild/buildtools",
							End: build.Position{
								Line:     6,
								LineRune: 45,
								Byte:     120,
							},
							Token: "'github.com/bazelbuild/buildtools'",
						},
					},
					&build.AssignExpr{
						LHS: &build.Ident{
							NamePos: build.Position{
								Line:     7,
								LineRune: 5,
								Byte:     126,
							},
							Name: "repo",
						},
						OpPos: build.Position{
							Line:     7,
							LineRune: 10,
							Byte:     131,
						},
						Op: "=",
						RHS: &build.StringExpr{
							Start: build.Position{
								Line:     7,
								LineRune: 12,
								Byte:     133,
							},
							Value: "github.com/bazelbuild/buildtools",
							End: build.Position{
								Line:     7,
								LineRune: 46,
								Byte:     167,
							},
							Token: "'github.com/bazelbuild/buildtools'",
						},
					},
					&build.AssignExpr{
						LHS: &build.Ident{
							NamePos: build.Position{
								Line:     8,
								LineRune: 5,
								Byte:     173,
							},
							Name: "revision",
						},
						OpPos: build.Position{
							Line:     8,
							LineRune: 14,
							Byte:     182,
						},
						Op: "=",
						RHS: &build.StringExpr{
							Start: build.Position{
								Line:     8,
								LineRune: 16,
								Byte:     184,
							},
							Value: "3.0.0",
							End: build.Position{
								Line:     8,
								LineRune: 23,
								Byte:     191,
							},
							Token: "'3.0.0'",
						},
					},
					&build.AssignExpr{
						LHS: &build.Ident{
							NamePos: build.Position{
								Line:     9,
								LineRune: 5,
								Byte:     197,
							},
							Name: "install",
						},
						OpPos: build.Position{
							Line:     9,
							LineRune: 13,
							Byte:     205,
						},
						Op: "=",
						RHS: &build.ListExpr{
							Start: build.Position{
								Line:     9,
								LineRune: 15,
								Byte:     207,
							},
							List: []build.Expr{
								&build.StringExpr{
									Start: build.Position{
										Line:     10,
										LineRune: 9,
										Byte:     217,
									},
									Value: "build",
									End: build.Position{
										Line:     10,
										LineRune: 16,
										Byte:     224,
									},
									Token: "'build'",
								},
								&build.StringExpr{
									Start: build.Position{
										Line:     11,
										LineRune: 9,
										Byte:     234,
									},
									Value: "tables",
									End: build.Position{
										Line:     11,
										LineRune: 17,
										Byte:     242,
									},
									Token: "'tables'",
								},
							},
							End: build.End{
								Pos: build.Position{
									Line:     12,
									LineRune: 5,
									Byte:     248,
								},
							},
						},
					},
				},
				End: build.End{
					Pos: build.Position{
						Line:     13,
						LineRune: 1,
						Byte:     251,
					},
				},
			},
		},
	}
}

func (suite *BuilderSuite) BuildtoolsRule(i int) build.Expr {
	return suite.Buildtools().Stmt[i]
}
