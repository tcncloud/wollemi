package filesystem

type Config struct {
	DefaultVisibility string            `json:"default_visibility,omitempty"`
	KnownDependency   map[string]string `json:"known_dependency,omitempty"`
}

func (this *Config) Merge(that *Config) *Config {
	if this == nil {
		return that
	}

	if that == nil {
		return this
	}

	size := (func() int {
		x := len(this.KnownDependency)
		y := len(that.KnownDependency)

		if x > y {
			return x
		}

		return y
	}())

	merge := &Config{
		DefaultVisibility: this.DefaultVisibility,
		KnownDependency:   make(map[string]string, size),
	}

	if that.DefaultVisibility != "" {
		merge.DefaultVisibility = that.DefaultVisibility
	}

	for key, value := range this.KnownDependency {
		merge.KnownDependency[key] = value
	}

	for key, value := range that.KnownDependency {
		merge.KnownDependency[key] = value
	}

	return merge
}
