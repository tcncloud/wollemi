package bazel

import (
	"fmt"

	"github.com/bazelbuild/buildtools/build"

	"github.com/tcncloud/wollemi/ports/please"
)

var encode Encode

type Encode struct{}

func (Encode) StringExpr(in *build.StringExpr) *please.StringExpr {
	if in == nil {
		return nil
	}

	return &please.StringExpr{Value: in.Value}
}

func (Encode) Ident(in *build.Ident) *please.Ident {
	if in == nil {
		return nil
	}

	return &please.Ident{Name: in.Name}
}

func (Encode) CallExpr(in *build.CallExpr) *please.CallExpr {
	if in == nil {
		return nil
	}

	out := &please.CallExpr{
		X: encode.Expr(in.X),
	}

	if in.List != nil {
		out.List = make([]please.Expr, len(in.List))

		for i, expr := range in.List {
			out.List[i] = encode.Expr(expr)
		}
	}

	return out
}

func (Encode) ListExpr(in *build.ListExpr) *please.ListExpr {
	if in == nil {
		return nil
	}

	out := &please.ListExpr{
		List: make([]please.Expr, len(in.List)),
	}

	for i, expr := range in.List {
		out.List[i] = encode.Expr(expr)
	}

	return out
}

func (Encode) AssignExpr(in *build.AssignExpr) *please.AssignExpr {
	if in == nil {
		return nil
	}

	return &please.AssignExpr{
		Op:  in.Op,
		LHS: encode.Expr(in.LHS),
		RHS: encode.Expr(in.RHS),
	}
}

func (Encode) BinaryExpr(in *build.BinaryExpr) *please.BinaryExpr {
	if in == nil {
		return nil
	}

	return &please.BinaryExpr{
		Op: in.Op,
		X:  encode.Expr(in.X),
		Y:  encode.Expr(in.Y),
	}
}

func (Encode) DictExpr(in *build.DictExpr) *please.DictExpr {
	if in == nil {
		return nil
	}

	out := &please.DictExpr{}

	if in.List != nil {
		out.List = make([]please.Expr, len(in.List))

		for i, expr := range in.List {
			out.List[i] = encode.Expr(expr)
		}
	}

	return out
}

func (Encode) KeyValueExpr(in *build.KeyValueExpr) *please.KeyValueExpr {
	if in == nil {
		return nil
	}

	return &please.KeyValueExpr{
		Key:   encode.Expr(in.Key),
		Value: encode.Expr(in.Value),
	}
}

func (Encode) LiteralExpr(in *build.LiteralExpr) *please.LiteralExpr {
	if in == nil {
		return nil
	}

	return &please.LiteralExpr{Token: in.Token}
}

func (Encode) Expr(in build.Expr) please.Expr {
	switch expr := in.(type) {
	case *build.Ident:
		return encode.Ident(expr)
	case *build.StringExpr:
		return encode.StringExpr(expr)
	case *build.ListExpr:
		return encode.ListExpr(expr)
	case *build.CallExpr:
		return encode.CallExpr(expr)
	case *build.AssignExpr:
		return encode.AssignExpr(expr)
	case *build.BinaryExpr:
		return encode.BinaryExpr(expr)
	case *build.DictExpr:
		return encode.DictExpr(expr)
	case *build.KeyValueExpr:
		return encode.KeyValueExpr(expr)
	case *build.LiteralExpr:
		return encode.LiteralExpr(expr)
	case nil:
		return nil
	default:
		panic(fmt.Errorf("unexpected bazel expr type %T", expr))
	}
}

func (Encode) Comments(in build.Comments) please.Comments {
	return please.Comments{
		Before: encode.CommentList(in.Before),
		Suffix: encode.CommentList(in.Suffix),
		After:  encode.CommentList(in.After),
	}
}

func (Encode) CommentList(in []build.Comment) []please.Comment {
	if in == nil {
		return nil
	}

	out := make([]please.Comment, len(in))

	for i, comment := range in {
		out[i] = encode.Comment(comment)
	}

	return out
}

func (Encode) Comment(in build.Comment) please.Comment {
	return please.Comment{Token: in.Token}
}
