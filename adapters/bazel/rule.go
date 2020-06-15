package bazel

import (
	"strings"

	"github.com/bazelbuild/buildtools/build"

	"github.com/tcncloud/wollemi/ports/please"
)

func NewRule(call *build.CallExpr) *Rule {
	return &Rule{Rule: build.NewRule(call)}
}

type Rule struct {
	*build.Rule
}

func (this *Rule) Unwrap() *build.Rule {
	return this.Rule
}

func (this *Rule) Comment() please.Comments {
	return encode.Comments(*this.Call.Comment())
}

func (this *Rule) Attr(key string) please.Expr {
	return encode.Expr(this.Rule.Attr(key))
}

func (this *Rule) AttrDefn(key string) *please.AssignExpr {
	return encode.AssignExpr(this.Rule.AttrDefn(key))
}

func (this *Rule) SetAttr(name string, val please.Expr) {
	switch val := val.(type) {
	case *please.ListExpr:
		var next *build.ListExpr

		switch attr := this.Rule.Attr(name).(type) {
		case *build.ListExpr:
			keep := make([]int, 0, len(attr.List))

			for i, entry := range attr.List {
				for _, suffix := range entry.Comment().Suffix {
					token := strings.TrimSpace(suffix.Token)

					if strings.EqualFold(token, "# wollemi:keep") {
						keep = append(keep, i)
						break
					}
				}
			}

			next = &build.ListExpr{
				List: make([]build.Expr, 0, len(keep)+len(val.List)),
			}

			for _, i := range keep {
				next.List = append(next.List, attr.List[i])
			}
		}

		if next == nil {
			next = &build.ListExpr{
				List: make([]build.Expr, 0, len(val.List)),
			}
		}

		for _, expr := range val.List {
			next.List = append(next.List, decode.Expr(expr))
		}

		this.Rule.SetAttr(name, next)
	default:
		this.Rule.SetAttr(name, decode.Expr(val))
	}
}

func (this *Rule) DelAttr(name string) please.Expr {
	return encode.Expr(this.Rule.DelAttr(name))
}
