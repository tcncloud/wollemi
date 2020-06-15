package please

type Rule interface {
	Attr(key string) Expr
	AttrDefn(key string) *AssignExpr
	AttrKeys() []string
	AttrLiteral(key string) string
	AttrString(key string) string
	AttrStrings(key string) []string
	DelAttr(key string) Expr
	ExplicitName() string
	Kind() string
	Name() string
	SetAttr(key string, val Expr)
	SetKind(kind string)
	Comment() Comments
}
