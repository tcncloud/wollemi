package please

type Graph struct {
	Packages map[string]*GraphPackage `json:"packages,omitempty"`
}

type GraphPackage struct {
	Targets map[string]*GraphTarget `json:"targets,omitempty"`
}

type GraphTarget struct {
	Inputs   []string `json:"inputs,omitempty"`
	Outs     []string `json:"outs,omitempty"`
	Deps     []string `json:"deps,omitempty"`
	Data     []string `json:"data,omitempty"`
	Labels   []string `json:"labels,omitempty"`
	Requires []string `json:"requires,omitempty"`
	Hash     string   `json:"hash,omitempty"`
	Binary   bool     `json:"binary,omitempty"`
}
