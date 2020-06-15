package please

type File interface {
	GetPath() string
	GetRules(func(Rule))
	GetRule(string) Rule
	SetRule(Rule)
	DelRule(string) bool
}
