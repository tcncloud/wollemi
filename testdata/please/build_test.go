package please_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/testdata/please"
)

func TestBuildFile_GetRule(t *testing.T) {
	t.Run("returns rule when it exists", func(t *testing.T) {
		file := &please.BuildFile{
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
				please.NewCallExpr("go_get", []please.Expr{
					please.NewAssignExpr("=", "name", "viper"),
					please.NewAssignExpr("=", "get", "github.com/spf13/pflag"),
					please.NewAssignExpr("=", "revision", "v1.6.3"),
				}),
			},
		}

		rule := file.GetRule("viper")
		require.NotNil(t, rule)
		require.Equal(t, "viper", rule.Name())
		require.Equal(t, "go_get", rule.Kind())
	})
}

func TestBuildFile_DelRule(t *testing.T) {
	t.Run("removes rule when it exists", func(t *testing.T) {
		file := &please.BuildFile{
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
				please.NewCallExpr("go_get", []please.Expr{
					please.NewAssignExpr("=", "name", "viper"),
					please.NewAssignExpr("=", "get", "github.com/spf13/pflag"),
					please.NewAssignExpr("=", "revision", "v1.6.3"),
				}),
			},
		}

		require.True(t, file.DelRule("viper"))
		require.Nil(t, file.GetRule("viper"))
	})
}
