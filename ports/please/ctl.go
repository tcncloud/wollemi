package please

type Ctl interface {
	QueryDeps(...string) ([]string, error)
	Graph() (*Graph, error)
	Build(...string) error
	Config(string) (Config, error)
}
