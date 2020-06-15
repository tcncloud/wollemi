package please

import (
	"fmt"
)

func NewIdent(name string) *Ident {
	return &Ident{Name: name}
}

func NewStringExpr(value string) *StringExpr {
	return &StringExpr{Value: value}
}

func NewListExpr(values ...interface{}) *ListExpr {
	if values == nil {
		return &ListExpr{}
	}

	out := &ListExpr{List: make([]Expr, len(values))}

	for i, value := range values {
		out.List[i] = NewExpr(value)
	}

	return out
}

func NewCallExpr(x string, list []Expr) *CallExpr {
	return &CallExpr{
		X:    NewIdent(x),
		List: list,
	}
}

func NewAssignExpr(op, lhs string, rhs interface{}) *AssignExpr {
	return &AssignExpr{
		Op:  op,
		LHS: NewIdent(lhs),
		RHS: NewExpr(rhs),
	}
}

func NewBinaryExpr(op string, x, y interface{}) *BinaryExpr {
	return &BinaryExpr{
		Op: op,
		X:  NewExpr(x),
		Y:  NewExpr(y),
	}
}

func NewGlob(include []string, exclude ...string) *CallExpr {
	call := NewCallExpr("glob", []Expr{NewExpr(include)})

	if len(exclude) > 0 {
		call.List = append(call.List, NewAssignExpr("=", "exclude", exclude))
	}

	return call
}

func NewExpr(value interface{}) Expr {
	switch v := value.(type) {
	case []Expr:
		return &ListExpr{List: v}
	case Expr:
		return v
	case bool:
		switch v {
		case true:
			return NewIdent("True")
		default:
			return NewIdent("False")
		}
	case string:
		return NewStringExpr(v)
	case []interface{}:
		return NewListExpr(v...)
	case []string:
		out := &ListExpr{List: make([]Expr, len(v))}

		for i, s := range v {
			out.List[i] = NewStringExpr(s)
		}

		return out
	default:
		panic(fmt.Errorf("unexpected type: %T", v))
	}
}
