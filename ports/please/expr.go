package please

type Ident struct {
	Name string
}

type StringExpr struct {
	Value string
}

type ListExpr struct {
	List []Expr
}

type CallExpr struct {
	X    Expr
	List []Expr
}

type AssignExpr struct {
	Op  string
	LHS Expr
	RHS Expr
}

type BinaryExpr struct {
	Op string
	X  Expr
	Y  Expr
}

type DictExpr struct {
	List []Expr
}

type KeyValueExpr struct {
	Key   Expr
	Value Expr
}

type LiteralExpr struct {
	Token string
}

// -----------------------------------------------------------------------------

type Expr interface{ isExpr() }

func (*Ident) isExpr()        {}
func (*StringExpr) isExpr()   {}
func (*ListExpr) isExpr()     {}
func (*CallExpr) isExpr()     {}
func (*AssignExpr) isExpr()   {}
func (*BinaryExpr) isExpr()   {}
func (*DictExpr) isExpr()     {}
func (*KeyValueExpr) isExpr() {}
func (*LiteralExpr) isExpr()  {}

// -----------------------------------------------------------------------------

func String(value string) *StringExpr {
	return &StringExpr{Value: value}
}

func Strings(values ...string) *ListExpr {
	out := &ListExpr{
		List: make([]Expr, len(values)),
	}

	for i, value := range values {
		out.List[i] = String(value)
	}

	return out
}

func Glob(include []string, exclude []string, targets ...string) Expr {
	glob := &CallExpr{
		X:    &Ident{Name: "glob"},
		List: make([]Expr, 0, 2),
	}

	glob.List = append(glob.List, Strings(include...))

	if len(exclude) > 0 {
		glob.List = append(glob.List, &AssignExpr{
			Op:  "=",
			LHS: &Ident{Name: "exclude"},
			RHS: Strings(exclude...),
		})
	}

	if len(targets) <= 0 {
		return glob
	}

	return &BinaryExpr{
		Op: "+",
		X:  glob,
		Y:  Strings(targets...),
	}
}

func Assign(name, op string, val Expr) *AssignExpr {
	return &AssignExpr{
		Op:  op,
		LHS: &Ident{Name: name},
		RHS: val,
	}
}

func Attr(expr *CallExpr, key string) Expr {
	for _, entry := range expr.List {
		switch assign := entry.(type) {
		case *AssignExpr:
			switch lhs := assign.LHS.(type) {
			case *Ident:
				if lhs.Name == key {
					return assign.RHS
				}
			}
		}
	}

	return nil
}

func AttrString(expr *CallExpr, key string) string {
	switch attr := Attr(expr, key).(type) {
	case *StringExpr:
		return attr.Value
	}

	return ""
}
