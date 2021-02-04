package bazel_test

import (
	"github.com/bazelbuild/buildtools/build"

	"github.com/tcncloud/wollemi/ports/please"
)

var factory = struct {
	build  BuildFactory
	please PleaseFactory
}{}

type BuildFactory struct{}

func (BuildFactory) StringExpr(value string) *build.StringExpr {
	return &build.StringExpr{Value: value}
}

func (BuildFactory) Ident(name string) *build.Ident {
	return &build.Ident{Name: name}
}

func (BuildFactory) CallExpr(kind string, attrs ...*build.AssignExpr) *build.CallExpr {
	call := &build.CallExpr{
		X:    factory.build.Ident(kind),
		List: make([]build.Expr, len(attrs)),
	}

	for i, attr := range attrs {
		call.List[i] = attr
	}

	return call
}

func (BuildFactory) ListExpr(list ...string) *build.ListExpr {
	out := &build.ListExpr{
		List: make([]build.Expr, len(list)),
	}

	for i, s := range list {
		out.List[i] = factory.build.StringExpr(s)
	}

	return out
}

func (BuildFactory) AssignExpr(name, op string, val build.Expr) *build.AssignExpr {
	return &build.AssignExpr{
		Op:  "=",
		LHS: factory.build.Ident(name),
		RHS: val,
	}
}

func (BuildFactory) BinaryExpr(x build.Expr, op string, y build.Expr) *build.BinaryExpr {
	return &build.BinaryExpr{X: x, Op: op, Y: y}
}

func (BuildFactory) DictExpr(kvs ...*build.KeyValueExpr) *build.DictExpr {
	out := &build.DictExpr{
		List: make([]*build.KeyValueExpr, len(kvs)),
	}

	for i, kv := range kvs {
		out.List[i] = kv
	}

	return out
}

func (BuildFactory) KeyValueExpr(key, value string) *build.KeyValueExpr {
	return &build.KeyValueExpr{
		Key:   factory.build.StringExpr(key),
		Value: factory.build.StringExpr(value),
	}
}

func (BuildFactory) LiteralExpr(token string) *build.LiteralExpr {
	return &build.LiteralExpr{Token: token}
}

func (BuildFactory) Comments() build.Comments {
	return build.Comments{
		Before: []build.Comment{{Token: "before"}},
		Suffix: []build.Comment{{Token: "suffix"}},
		After:  []build.Comment{{Token: "after"}},
	}
}

// -----------------------------------------------------------------------------

type PleaseFactory struct{}

func (PleaseFactory) StringExpr(value string) *please.StringExpr {
	return &please.StringExpr{Value: value}
}

func (PleaseFactory) Ident(name string) *please.Ident {
	return &please.Ident{Name: name}
}

func (PleaseFactory) CallExpr(kind string, attrs ...*please.AssignExpr) *please.CallExpr {
	call := &please.CallExpr{
		X:    factory.please.Ident(kind),
		List: make([]please.Expr, len(attrs)),
	}

	for i, attr := range attrs {
		call.List[i] = attr
	}

	return call
}

func (PleaseFactory) ListExpr(list ...string) *please.ListExpr {
	out := &please.ListExpr{
		List: make([]please.Expr, len(list)),
	}

	for i, s := range list {
		out.List[i] = factory.please.StringExpr(s)
	}

	return out
}

func (PleaseFactory) AssignExpr(name, op string, val please.Expr) *please.AssignExpr {
	return &please.AssignExpr{
		Op:  "=",
		LHS: factory.please.Ident(name),
		RHS: val,
	}
}

func (PleaseFactory) BinaryExpr(x please.Expr, op string, y please.Expr) *please.BinaryExpr {
	return &please.BinaryExpr{X: x, Op: op, Y: y}
}

func (PleaseFactory) DictExpr(kvs ...*please.KeyValueExpr) *please.DictExpr {
	out := &please.DictExpr{
		List: make([]please.Expr, len(kvs)),
	}

	for i, kv := range kvs {
		out.List[i] = kv
	}

	return out
}

func (PleaseFactory) KeyValueExpr(key, value string) *please.KeyValueExpr {
	return &please.KeyValueExpr{
		Key:   factory.please.StringExpr(key),
		Value: factory.please.StringExpr(value),
	}
}

func (PleaseFactory) LiteralExpr(token string) *please.LiteralExpr {
	return &please.LiteralExpr{Token: token}
}

func (PleaseFactory) Comments() please.Comments {
	return please.Comments{
		Before: []please.Comment{{Token: "before"}},
		Suffix: []please.Comment{{Token: "suffix"}},
		After:  []please.Comment{{Token: "after"}},
	}
}
