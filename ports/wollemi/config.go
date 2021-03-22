package wollemi

import (
	"encoding/json"
	"strconv"

	"github.com/tcncloud/wollemi/domain/optional"
)

func Bool(value bool) *bool { return &value }

type Config struct {
	Gofmt                     Gofmt             `json:"gofmt,omitempty"`
	DefaultVisibility         string            `json:"default_visibility,omitempty"`
	KnownDependency           map[string]string `json:"known_dependency,omitempty"`
	AllowUnresolvedDependency *optional.Bool    `json:"allow_unresolved_dependency,omitempty"`
	ExplicitSources           *optional.Bool    `json:"explicit_sources,omitempty"`
}

type Gofmt struct {
	Rewrite *bool  `json:"rewrite,omitempty"`
	Create  create `json:"create,omitempty"`
}

func (gofmt *Gofmt) GetCreate() []string {
	if gofmt != nil && gofmt.Create != nil {
		return gofmt.Create
	}

	return []string{"go_binary", "go_library", "go_test"}
}

func (gofmt *Gofmt) GetRewrite() bool {
	if gofmt != nil && gofmt.Rewrite != nil {
		return *gofmt.Rewrite
	}

	return true
}

func (this Config) Merge(that Config) Config {
	merge := this

	if that.DefaultVisibility != "" {
		merge.DefaultVisibility = that.DefaultVisibility
	}

	if len(this.KnownDependency) > 0 || len(that.KnownDependency) > 0 {
		size := (func() int {
			x := len(this.KnownDependency)
			y := len(that.KnownDependency)

			if x > y {
				return x
			}

			return y
		}())

		merge.KnownDependency = make(map[string]string, size)

		for key, value := range this.KnownDependency {
			merge.KnownDependency[key] = value
		}

		for key, value := range that.KnownDependency {
			merge.KnownDependency[key] = value
		}
	}

	if v := that.AllowUnresolvedDependency; v != nil {
		merge.AllowUnresolvedDependency = v
	}

	if v := that.ExplicitSources; v != nil {
		merge.ExplicitSources = v
	}

	if v := that.Gofmt.Rewrite; v != nil {
		merge.Gofmt.Rewrite = v
	}

	if v := that.Gofmt.Create; v != nil {
		merge.Gofmt.Create = v
	}

	return merge
}

type create []string

func (list *create) UnmarshalJSON(buf []byte) error {
	err := json.Unmarshal(buf, (*[]string)(list))
	if err != nil {
		s, err := strconv.Unquote(string(buf))
		if err == nil {
			switch s {
			case "on", "default":
				*list = (*Gofmt)(nil).GetCreate()
			case "off":
				*list = []string{}
			default:
			}
		}
	}

	return nil
}
