package please

type Graph struct {
	Packages map[string]*Package `json:"packages,omitempty"`
}

type Package struct {
	Targets map[string]*Target `json:"targets,omitempty"`
}

type Target struct {
	Inputs   []string `json:"inputs,omitempty"`
	Outs     []string `json:"outs,omitempty"`
	Srcs     []string `json:"srcs,omitempty"`
	Deps     []string `json:"deps,omitempty"`
	Data     []string `json:"data,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	Requires []string `json:"requires,omitempty"`
	Hash     string   `json:"hash,omitempty"`
	Binary   bool     `json:"binary,omitempty"`
}
