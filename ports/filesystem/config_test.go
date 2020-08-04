package filesystem_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/domain/optional"
	"github.com/tcncloud/wollemi/ports/filesystem"
)

func TestConfig_Merge(t *testing.T) {
	for _, tt := range []struct {
		Name string
		Lhs  *filesystem.Config
		Rhs  *filesystem.Config
		Want *filesystem.Config
	}{{
		Name: "merged config is lhs when rhs nil",
		Lhs: &filesystem.Config{
			DefaultVisibility:         "PUBLIC",
			AllowUnresolvedDependency: optional.BoolValue(true),
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
				"eee": "fff",
			},
		},
		Want: &filesystem.Config{
			DefaultVisibility:         "PUBLIC",
			AllowUnresolvedDependency: optional.BoolValue(true),
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
				"eee": "fff",
			},
		},
	}, {
		Name: "merged config is rhs when lhs nil",
		Rhs: &filesystem.Config{
			DefaultVisibility:         "PUBLIC",
			AllowUnresolvedDependency: optional.BoolValue(true),
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
				"eee": "fff",
			},
		},
		Want: &filesystem.Config{
			DefaultVisibility:         "PUBLIC",
			AllowUnresolvedDependency: optional.BoolValue(true),
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
				"eee": "fff",
			},
		},
	}, {
		Name: "merged default_visibility is lhs when rhs empty",
		Lhs: &filesystem.Config{
			DefaultVisibility: "PUBLIC",
		},
		Rhs: &filesystem.Config{},
		Want: &filesystem.Config{
			DefaultVisibility: "PUBLIC",
		},
	}, {
		Name: "merged default_visibility is rhs when rhs non empty",
		Lhs: &filesystem.Config{
			DefaultVisibility: "PUBLIC",
		},
		Rhs: &filesystem.Config{
			DefaultVisibility: "//app/...",
		},
		Want: &filesystem.Config{
			DefaultVisibility: "//app/...",
		},
	}, {
		Name: "merged allow_unresolved_dependency is rhs when rhs set",
		Lhs:  &filesystem.Config{},
		Rhs: &filesystem.Config{
			AllowUnresolvedDependency: optional.BoolValue(true),
		},
		Want: &filesystem.Config{
			AllowUnresolvedDependency: optional.BoolValue(true),
		},
	}, {
		Name: "merged allow_unresolved_dependency is lhs when rhs unset",
		Lhs: &filesystem.Config{
			AllowUnresolvedDependency: optional.BoolValue(true),
		},
		Rhs: &filesystem.Config{},
		Want: &filesystem.Config{
			AllowUnresolvedDependency: optional.BoolValue(true),
		},
	}, {
		Name: "merged known_dependency is all key values from rhs applied to lhs",
		Lhs: &filesystem.Config{
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "zzz",
			},
		},
		Rhs: &filesystem.Config{
			KnownDependency: map[string]string{
				"ccc": "ddd",
				"eee": "fff",
			},
		},
		Want: &filesystem.Config{
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
				"eee": "fff",
			},
		},
	}} {
		t.Run(tt.Name, func(t *testing.T) {
			require.Equal(t, tt.Want, tt.Lhs.Merge(tt.Rhs))
		})
	}
}
