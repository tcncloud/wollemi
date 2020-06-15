package please

import (
	"fmt"

	"github.com/tcncloud/wollemi/ports/please"
)

var cp Copier

type Copier struct{}

func (Copier) BuildFile(in *BuildFile) *BuildFile {
	if in == nil {
		return nil
	}

	out := &BuildFile{Path: in.Path}

	if in.Stmt != nil {
		out.Stmt = make([]please.Expr, len(in.Stmt))
	}

	for i, stmt := range in.Stmt {
		out.Stmt[i] = cp.Expr(stmt)
	}

	return out
}

func (Copier) Exprs(in []please.Expr) []please.Expr {
	if in == nil {
		return nil
	}

	out := make([]please.Expr, len(in))

	for i, expr := range in {
		out[i] = cp.Expr(expr)
	}

	return out
}

func (Copier) Expr(in please.Expr) please.Expr {
	switch in := in.(type) {
	case *please.Ident:
		return cp.Ident(in)
	case *please.StringExpr:
		return cp.StringExpr(in)
	case *please.ListExpr:
		return cp.ListExpr(in)
	case *please.CallExpr:
		return cp.CallExpr(in)
	case *please.AssignExpr:
		return cp.AssignExpr(in)
	case *please.BinaryExpr:
		return cp.BinaryExpr(in)
	case *please.DictExpr:
		return cp.DictExpr(in)
	case *please.KeyValueExpr:
		return cp.KeyValueExpr(in)
	case *please.LiteralExpr:
		return cp.LiteralExpr(in)
	case nil:
		return nil
	default:
		panic(fmt.Errorf("unexpected expr type %T", in))
	}
}

func (Copier) Ident(in *please.Ident) *please.Ident {
	if in == nil {
		return nil
	}

	return &please.Ident{
		Name: in.Name,
	}
}

func (Copier) StringExpr(in *please.StringExpr) *please.StringExpr {
	if in == nil {
		return nil
	}

	return &please.StringExpr{
		Value: in.Value,
	}
}

func (Copier) ListExpr(in *please.ListExpr) *please.ListExpr {
	if in == nil {
		return nil
	}

	return &please.ListExpr{
		List: cp.Exprs(in.List),
	}
}

func (Copier) CallExpr(in *please.CallExpr) *please.CallExpr {
	if in == nil {
		return nil
	}

	return &please.CallExpr{
		List: cp.Exprs(in.List),
		X:    cp.Expr(in.X),
	}
}

func (Copier) AssignExpr(in *please.AssignExpr) *please.AssignExpr {
	if in == nil {
		return nil
	}

	return &please.AssignExpr{
		LHS: cp.Expr(in.LHS),
		RHS: cp.Expr(in.RHS),
		Op:  in.Op,
	}
}

func (Copier) BinaryExpr(in *please.BinaryExpr) *please.BinaryExpr {
	if in == nil {
		return nil
	}

	return &please.BinaryExpr{
		X:  cp.Expr(in.X),
		Y:  cp.Expr(in.Y),
		Op: in.Op,
	}
}

func (Copier) DictExpr(in *please.DictExpr) *please.DictExpr {
	if in == nil {
		return nil
	}

	return &please.DictExpr{
		List: cp.Exprs(in.List),
	}
}

func (Copier) KeyValueExpr(in *please.KeyValueExpr) *please.KeyValueExpr {
	if in == nil {
		return nil
	}

	return &please.KeyValueExpr{
		Value: cp.Expr(in.Value),
		Key:   cp.Expr(in.Key),
	}
}

func (Copier) LiteralExpr(in *please.LiteralExpr) *please.LiteralExpr {
	if in == nil {
		return nil
	}

	return &please.LiteralExpr{
		Token: in.Token,
	}
}
