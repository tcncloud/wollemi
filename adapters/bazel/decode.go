package bazel

import (
	"fmt"

	"github.com/bazelbuild/buildtools/build"

	"github.com/tcncloud/wollemi/ports/please"
)

var decode Decode

type Decode struct{}

func (Decode) StringExpr(in *please.StringExpr) *build.StringExpr {
	if in == nil {
		return nil
	}

	return &build.StringExpr{Value: in.Value}
}

func (Decode) Ident(in *please.Ident) *build.Ident {
	if in == nil {
		return nil
	}

	return &build.Ident{Name: in.Name}
}

func (Decode) CallExpr(in *please.CallExpr) *build.CallExpr {
	if in == nil {
		return nil
	}

	out := &build.CallExpr{
		X: decode.Expr(in.X),
	}

	if in.List != nil {
		out.List = make([]build.Expr, len(in.List))

		for i, expr := range in.List {
			out.List[i] = decode.Expr(expr)
		}
	}

	return out
}

func (Decode) ListExpr(in *please.ListExpr) *build.ListExpr {
	if in == nil {
		return nil
	}

	out := &build.ListExpr{
		List: make([]build.Expr, len(in.List)),
	}

	for i, expr := range in.List {
		out.List[i] = decode.Expr(expr)
	}

	return out
}

func (Decode) AssignExpr(in *please.AssignExpr) *build.AssignExpr {
	if in == nil {
		return nil
	}

	return &build.AssignExpr{
		Op:  in.Op,
		LHS: decode.Expr(in.LHS),
		RHS: decode.Expr(in.RHS),
	}
}

func (Decode) BinaryExpr(in *please.BinaryExpr) *build.BinaryExpr {
	if in == nil {
		return nil
	}

	return &build.BinaryExpr{
		Op: in.Op,
		X:  decode.Expr(in.X),
		Y:  decode.Expr(in.Y),
	}
}

func (Decode) DictExpr(in *please.DictExpr) *build.DictExpr {
	if in == nil {
		return nil
	}

	out := &build.DictExpr{}

	if in.List != nil {
		out.List = make([]*build.KeyValueExpr, len(in.List))

		for i, expr := range in.List {
			// TODO: rework the please type to reflect the buildtools type
			out.List[i] = decode.Expr(expr).(*build.KeyValueExpr)
		}
	}

	return out
}

func (Decode) KeyValueExpr(in *please.KeyValueExpr) *build.KeyValueExpr {
	if in == nil {
		return nil
	}

	return &build.KeyValueExpr{
		Key:   decode.Expr(in.Key),
		Value: decode.Expr(in.Value),
	}
}

func (Decode) LiteralExpr(in *please.LiteralExpr) *build.LiteralExpr {
	if in == nil {
		return nil
	}

	return &build.LiteralExpr{Token: in.Token}
}

func (Decode) Expr(in please.Expr) build.Expr {
	switch expr := in.(type) {
	case *please.Ident:
		return decode.Ident(expr)
	case *please.StringExpr:
		return decode.StringExpr(expr)
	case *please.ListExpr:
		return decode.ListExpr(expr)
	case *please.CallExpr:
		return decode.CallExpr(expr)
	case *please.AssignExpr:
		return decode.AssignExpr(expr)
	case *please.BinaryExpr:
		return decode.BinaryExpr(expr)
	case *please.DictExpr:
		return decode.DictExpr(expr)
	case *please.KeyValueExpr:
		return decode.KeyValueExpr(expr)
	case *please.LiteralExpr:
		return decode.LiteralExpr(expr)
	case nil:
		return nil
	default:
		panic(fmt.Errorf("unexpected please expr type %T", expr))
	}
}
