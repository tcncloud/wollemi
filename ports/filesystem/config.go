package filesystem

import (
	"github.com/tcncloud/wollemi/domain/optional"
)

type Config struct {
	DefaultVisibility         string            `json:"default_visibility,omitempty"`
	KnownDependency           map[string]string `json:"known_dependency,omitempty"`
	AllowUnresolvedDependency *optional.Bool    `json:"allow_unresolved_dependency,omitempty"`
}

func (this *Config) Merge(that *Config) *Config {
	if this == nil {
		return that
	}

	if that == nil {
		return this
	}

	merge := &Config{
		DefaultVisibility:         this.DefaultVisibility,
		AllowUnresolvedDependency: this.AllowUnresolvedDependency,
	}

	if that.DefaultVisibility != "" {
		merge.DefaultVisibility = that.DefaultVisibility
	}

	if len(this.KnownDependency) > 0 {
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
	}

	for key, value := range that.KnownDependency {
		merge.KnownDependency[key] = value
	}

	if v := that.AllowUnresolvedDependency; v != nil {
		merge.AllowUnresolvedDependency = v
	}

	return merge
}
