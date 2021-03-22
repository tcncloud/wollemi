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
	Manage  manage `json:"manage,omitempty"`
}

func (gofmt *Gofmt) GetRewrite() bool {
	if gofmt != nil && gofmt.Rewrite != nil {
		return *gofmt.Rewrite
	}

	return true
}

func (gofmt *Gofmt) GetCreate() []string {
	if gofmt != nil && gofmt.Create != nil {
		return gofmt.Create
	}

	return []string{"go_binary", "go_library", "go_test"}
}

func (gofmt *Gofmt) GetManage() []string {
	if gofmt != nil && gofmt.Manage != nil {
		return gofmt.Manage
	}

	return []string{"go_binary", "go_library", "go_test"}
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

	if v := that.Gofmt.Manage; v != nil {
		merge.Gofmt.Manage = v
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

type manage []string

func (list *manage) UnmarshalJSON(buf []byte) error {
	var gofmt *Gofmt

	err := json.Unmarshal(buf, (*[]string)(list))

	if err != nil {
		s, unquoteErr := strconv.Unquote(string(buf))
		if unquoteErr == nil {
			switch s {
			case "on", "default":
				*list = gofmt.GetManage()
				err = nil // recover
			case "off":
				*list = []string{}
				err = nil // recover
			}
		}
	}

	if err == nil && len(*list) > 0 {
		var expand []string

		for _, kind := range *list {
			if kind == "default" {
				expand = appendUniqString(expand, gofmt.GetManage()...)
			} else {
				expand = appendUniqString(expand, kind)
			}
		}

		*list = expand
	}

	return nil
}

func inStrings(from []string, value string) bool {
	return indexStrings(from, value) >= 0
}

func appendUniqString(dest []string, from ...string) []string {
	for _, s := range from {
		if !inStrings(dest, s) {
			dest = append(dest, s)
		}
	}

	return dest
}

func indexStrings(from []string, value string) int {
	for i, have := range from {
		if value == have {
			return i
		}
	}

	return -1
}
