package wollemi_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/domain/optional"
	"github.com/tcncloud/wollemi/ports/wollemi"
)

func TestConfig_UnmarshalJSON(t *testing.T) {
	for _, tt := range []struct {
		Title string
		Data  string
		Want  wollemi.Config
	}{{
		Title: "unmarshals simple json config",
		Want: wollemi.Config{
			DefaultVisibility:         "//project/service/routes/...",
			AllowUnresolvedDependency: optional.BoolValue(true),
			ExplicitSources:           optional.BoolValue(true),
			KnownDependency: map[string]string{
				"github.com/olivere/elastic": "//third_party/go/github.com/olivere/elastic:v7",
			},
			Gofmt: wollemi.Gofmt{
				Rewrite: wollemi.Bool(true),
				Create:  []string{"go_library", "go_test"},
				Manage:  []string{"go_binary", "go_test"},
				Mapped: map[string]string{
					"go_binary":  "go_custom_binary",
					"go_library": "go_library",
					"go_test":    "go_custom_test",
				},
			},
		},
		Data: `{
      "default_visibility": "//project/service/routes/...",
      "allow_unresolved_dependency": true,
      "explicit_sources": true,
      "known_dependency": {
        "github.com/olivere/elastic": "//third_party/go/github.com/olivere/elastic:v7"
      },
      "gofmt": {
        "rewrite": true,
        "create": ["go_library", "go_test"],
        "manage": ["go_binary", "go_test"],
        "mapped": {
          "go_binary": "go_custom_binary",
          "go_test": "go_custom_test"
        }
      }
    }`,
	}, {
		Title: "unmarshals json config when gofmt create set to on",
		Data:  `{"gofmt":{"create":"on"}}`,
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Create: []string{"go_binary", "go_library", "go_test"},
			},
		},
	}, {
		Title: "unmarshals json config when gofmt create set to default",
		Data:  `{"gofmt":{"create":"default"}}`,
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Create: []string{"go_binary", "go_library", "go_test"},
			},
		},
	}, {
		Title: "unmarshals json config when gofmt create set to off",
		Data:  `{"gofmt":{"create":"off"}}`,
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Create: []string{},
			},
		},
	}, {
		Title: "unmarshals json config when gofmt manage set to on",
		Data:  `{"gofmt":{"manage":"on"}}`,
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Manage: []string{"go_binary", "go_library", "go_test"},
			},
		},
	}, {
		Title: "unmarshals json config when gofmt manage set to default",
		Data:  `{"gofmt":{"manage":"default"}}`,
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Manage: []string{"go_binary", "go_library", "go_test"},
			},
		},
	}, {
		Title: "unmarshals json config when gofmt manage set to off",
		Data:  `{"gofmt":{"manage":"off"}}`,
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Manage: []string{},
			},
		},
	}, {
		Title: "unmarshals json config when gofmt manage contains default",
		Data:  `{"gofmt":{"manage":["default", "go_custom_binary"]}}`,
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Manage: []string{"go_binary", "go_library", "go_test", "go_custom_binary"},
			},
		},
	}, {
		Title: "unmarshals json config when gofmt mapped set to none",
		Data:  `{"gofmt":{"mapped":"none"}}`,
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Mapped: map[string]string{
					"go_binary":  "go_binary",
					"go_library": "go_library",
					"go_test":    "go_test",
				},
			},
		},
	}} {
		t.Run(tt.Title, func(t *testing.T) {
			have := wollemi.Config{}

			err := json.Unmarshal([]byte(tt.Data), &have)
			require.NoError(t, err)

			require.Equal(t, tt.Want, have)
		})
	}
}

func TestConfig_Merge(t *testing.T) {
	for _, tt := range []struct {
		Name string
		Lhs  wollemi.Config
		Rhs  wollemi.Config
		Want wollemi.Config
	}{{
		Name: "merged config is lhs when rhs empty",
		Lhs: wollemi.Config{
			DefaultVisibility:         "PUBLIC",
			AllowUnresolvedDependency: optional.BoolValue(true),
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
				"eee": "fff",
			},
		},
		Want: wollemi.Config{
			DefaultVisibility:         "PUBLIC",
			AllowUnresolvedDependency: optional.BoolValue(true),
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
				"eee": "fff",
			},
		},
	}, {
		Name: "merged config is rhs when lhs empty",
		Rhs: wollemi.Config{
			DefaultVisibility:         "PUBLIC",
			AllowUnresolvedDependency: optional.BoolValue(true),
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
				"eee": "fff",
			},
		},
		Want: wollemi.Config{
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
		Lhs: wollemi.Config{
			DefaultVisibility: "PUBLIC",
		},
		Rhs: wollemi.Config{},
		Want: wollemi.Config{
			DefaultVisibility: "PUBLIC",
		},
	}, {
		Name: "merged default_visibility is rhs when rhs non empty",
		Lhs: wollemi.Config{
			DefaultVisibility: "PUBLIC",
		},
		Rhs: wollemi.Config{
			DefaultVisibility: "//app/...",
		},
		Want: wollemi.Config{
			DefaultVisibility: "//app/...",
		},
	}, {
		Name: "merged allow_unresolved_dependency is rhs when rhs set",
		Lhs:  wollemi.Config{},
		Rhs: wollemi.Config{
			AllowUnresolvedDependency: optional.BoolValue(true),
		},
		Want: wollemi.Config{
			AllowUnresolvedDependency: optional.BoolValue(true),
		},
	}, {
		Name: "merged allow_unresolved_dependency is lhs when rhs unset",
		Lhs: wollemi.Config{
			AllowUnresolvedDependency: optional.BoolValue(true),
		},
		Rhs: wollemi.Config{},
		Want: wollemi.Config{
			AllowUnresolvedDependency: optional.BoolValue(true),
		},
	}, {
		Name: "merged known_dependency is all key values from rhs applied to lhs",
		Lhs: wollemi.Config{
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "zzz",
			},
		},
		Rhs: wollemi.Config{
			KnownDependency: map[string]string{
				"ccc": "ddd",
				"eee": "fff",
			},
		},
		Want: wollemi.Config{
			KnownDependency: map[string]string{
				"aaa": "bbb",
				"ccc": "ddd",
				"eee": "fff",
			},
		},
	}, {
		Name: "merged gofmt mapped is lhs when rhs is null",
		Lhs: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Mapped: map[string]string{
					"go_library": "go_library",
				},
			},
		},
		Rhs: wollemi.Config{},
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Mapped: map[string]string{
					"go_library": "go_library",
				},
			},
		},
	}, {
		Name: "merged gofmt mapped is rhs when rhs is non null",
		Lhs: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Mapped: map[string]string{
					"go_library": "go_library",
				},
			},
		},
		Rhs: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Mapped: map[string]string{
					"go_binary": "go_binary",
				},
			},
		},
		Want: wollemi.Config{
			Gofmt: wollemi.Gofmt{
				Mapped: map[string]string{
					"go_binary": "go_binary",
				},
			},
		},
	}} {
		t.Run(tt.Name, func(t *testing.T) {
			require.Equal(t, tt.Want, tt.Lhs.Merge(tt.Rhs))
		})
	}
}
