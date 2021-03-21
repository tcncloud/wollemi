package please

import (
	"fmt"

	"github.com/tcncloud/wollemi/ports/please"
)

type (
	CallExpr     = please.CallExpr
	AssignExpr   = please.AssignExpr
	StringExpr   = please.StringExpr
	ListExpr     = please.ListExpr
	BinaryExpr   = please.BinaryExpr
	Ident        = please.Ident
	Expr         = please.Expr
	File         = please.File
	DictExpr     = please.DictExpr
	KeyValueExpr = please.KeyValueExpr
	LiteralExpr  = please.LiteralExpr
	Graph        = please.Graph
	GraphTarget  = please.GraphTarget
	GraphPackage = please.GraphPackage
)

type BuildFile struct {
	Path string
	Stmt []please.Expr
}

func (this *BuildFile) GetPath() string {
	return this.Path
}

func (this *BuildFile) DelRule(name string) bool {
	for i, stmt := range this.Stmt {
		switch call := stmt.(type) {
		case *please.CallExpr:
			for _, expr := range call.List {
				switch assign := expr.(type) {
				case *please.AssignExpr:
					switch lhs := assign.LHS.(type) {
					case *please.Ident:
						if lhs.Name == "name" {
							switch rhs := assign.RHS.(type) {
							case *please.StringExpr:
								if rhs.Value == name {
									if i+1 < len(this.Stmt) {
										copy(this.Stmt[i:], this.Stmt[i+1:])
									}

									this.Stmt = this.Stmt[:len(this.Stmt)-1]

									return true
								}
							}
						}
					}
				}
			}
		}
	}

	return false
}

func (this *BuildFile) GetRule(name string) please.Rule {
	for _, stmt := range this.Stmt {
		switch call := stmt.(type) {
		case *please.CallExpr:
			for _, expr := range call.List {
				switch assign := expr.(type) {
				case *please.AssignExpr:
					switch lhs := assign.LHS.(type) {
					case *please.Ident:
						if lhs.Name == "name" {
							switch rhs := assign.RHS.(type) {
							case *please.StringExpr:
								if rhs.Value == name {
									return &Rule{Call: call}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (this *BuildFile) GetRules(yield func(please.Rule)) {
	for _, stmt := range this.Stmt {
		switch call := stmt.(type) {
		case *please.CallExpr:
			yield(&Rule{Call: call})
		}
	}
}

func (this *BuildFile) SetRule(rule please.Rule) {
	for i, stmt := range this.Stmt {
		switch call := stmt.(type) {
		case *please.CallExpr:
			have := &Rule{Call: call}

			if have.Name() == rule.Name() {
				switch rule := rule.(type) {
				case *Rule:
					this.Stmt[i] = rule.Call
					return
				default:
					panic(fmt.Errorf("rule not created by package"))
				}
			}
		}
	}

	switch rule := rule.(type) {
	case *Rule:
		this.Stmt = append(this.Stmt, rule.Call)
	default:
		panic(fmt.Errorf("rule not created by package"))
	}
}

func NewRule(kind, name string) *Rule {
	return &Rule{
		Call: &please.CallExpr{
			X: &please.Ident{Name: kind},
			List: []please.Expr{
				please.Assign("name", "=", please.String(name)),
			},
		},
	}
}

type Rule struct {
	Call *please.CallExpr
}

func (this *Rule) Unwrap() *please.CallExpr {
	return this.Call
}

func (this *Rule) Attr(key string) please.Expr {
	for _, attr := range this.Call.List {
		switch assign := attr.(type) {
		case *please.AssignExpr:
			switch lhs := assign.LHS.(type) {
			case *please.Ident:
				if lhs.Name == key {
					return assign.RHS
				}
			}
		}
	}

	return nil
}

func (this *Rule) AttrDefn(name string) *please.AssignExpr {
	return nil
}

func (this *Rule) AttrKeys() []string {
	keys := make([]string, 0, len(this.Call.List))

	for _, attr := range this.Call.List {
		switch assign := attr.(type) {
		case *please.AssignExpr:
			switch lhs := assign.LHS.(type) {
			case *please.Ident:
				keys = append(keys, lhs.Name)
			}
		}
	}

	return keys
}

func (this *Rule) AttrLiteral(key string) string {
	switch attr := this.Attr(key).(type) {
	case *please.Ident:
		return attr.Name
	}

	return ""
}

func (this *Rule) AttrString(key string) string {
	switch attr := this.Attr(key).(type) {
	case *please.StringExpr:
		return attr.Value
	}

	return ""
}

func (this *Rule) AttrStrings(key string) []string {
	switch attr := this.Attr(key).(type) {
	case *please.ListExpr:
		if attr.List == nil {
			return nil
		}

		out := make([]string, 0, len(attr.List))

		for _, entry := range attr.List {
			switch expr := entry.(type) {
			case *please.StringExpr:
				out = append(out, expr.Value)
			default:
				return nil
			}
		}

		return out
	}

	return nil
}

func (this *Rule) DelAttr(key string) please.Expr {
	for i, attr := range this.Call.List {
		switch assign := attr.(type) {
		case *please.AssignExpr:
			switch lhs := assign.LHS.(type) {
			case *please.Ident:
				if lhs.Name == key {
					if i+1 < len(this.Call.List) {
						copy(this.Call.List[i:], this.Call.List[i+1:])
					}

					this.Call.List = this.Call.List[:len(this.Call.List)-1]

					return assign.RHS
				}
			}
		}
	}

	return nil
}

func (this *Rule) ExplicitName() string {
	return ""
}

func (this *Rule) Kind() string {
	switch x := this.Call.X.(type) {
	case *please.Ident:
		return x.Name
	}

	return ""
}

func (this *Rule) SetKind(kind string) {
	this.Call.X = &please.Ident{Name: kind}
}

func (this *Rule) Name() string {
	return this.AttrString("name")
}

func (this *Rule) SetAttr(key string, val please.Expr) {
	for _, attr := range this.Call.List {
		switch assign := attr.(type) {
		case *please.AssignExpr:
			switch lhs := assign.LHS.(type) {
			case *please.Ident:
				if lhs.Name == key {
					assign.RHS = val
					return
				}
			}
		}
	}

	this.Call.List = append(this.Call.List, please.Assign(key, "=", val))
}

func (this *Rule) Comment() please.Comments {
	return please.Comments{}
}
