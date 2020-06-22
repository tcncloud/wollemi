package bazel

import (
	"fmt"

	"github.com/bazelbuild/buildtools/build"

	"github.com/tcncloud/wollemi/ports/please"
)

func NewFile(data []byte, file *build.File) *File {
	out := &File{
		Data: make([]byte, len(data)),
		File: file,
	}

	copy(out.Data, data)

	return out
}

type File struct {
	*build.File
	Data []byte
}

func (this *File) IsEmpty() bool {
	for _, stmt := range this.Stmt {
		switch expr := stmt.(type) {
		case *build.ForStmt:
			return false
		case *build.IfStmt:
			return false
		case *build.CallExpr:
			rule := NewRule(expr)
			switch rule.Kind() {
			case "package":
			case "subinclude":
			default:
				return false
			}
		}
	}

	return true
}

func (this *File) Unwrap() *build.File {
	return this.File
}

func (this *File) GetPath() string {
	return this.Path
}

func (this *File) GetRule(name string) please.Rule {
	for _, stmt := range this.Stmt {
		switch call := stmt.(type) {
		case *build.CallExpr:
			rule := NewRule(call)

			if rule.Name() == name {
				return rule
			}
		}
	}

	return nil
}

func (this *File) GetRules(yield func(please.Rule)) {
	for _, stmt := range this.Stmt {
		switch call := stmt.(type) {
		case *build.CallExpr:
			rule := NewRule(call)

			yield(rule)
		}
	}
}

func (this *File) SetRule(rule please.Rule) {
	name := rule.Name()

	var i int

Stmt:
	for ; i < len(this.Stmt); i++ {
		switch call := this.Stmt[i].(type) {
		case *build.CallExpr:
			if NewRule(call).Name() == name {
				break Stmt
			}
		}
	}

	switch rule := rule.(type) {
	case *Rule:
		if i >= len(this.Stmt) {
			this.Stmt = append(this.Stmt, rule.Call)
		} else {
			this.Stmt[i] = rule.Call
		}
	default:
		panic(fmt.Errorf("rule not created by package"))
	}
}

func (this *File) DelRule(name string) bool {
	rule := this.GetRule(name)
	if rule == nil {
		return false
	}

	return this.DelRules("", name) > 0
}
