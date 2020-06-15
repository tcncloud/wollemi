package stringify

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode"
)

func String(v interface{}, indent int) string {
	buf := bytes.NewBuffer(nil)

	Write(&writer{Writer: buf}, v, indent)

	return buf.String()
}

func Write(w io.Writer, value interface{}, indent int) error {
	var inner func(reflect.Value, bool)

	buf := &writer{Writer: w, indent: indent}

	var nl string

	if buf.indent >= 0 {
		nl = "\n"
	}

	inner = func(v reflect.Value, slice bool) {
		if v.IsZero() {
			return
		}

		t := v.Type()

		if t.Kind() == reflect.Interface {
			v = v.Elem()
			t = v.Type()
		}

		var isPtr bool

		if t.Kind() == reflect.Ptr {
			isPtr = true
			t = t.Elem()
			v = v.Elem()
		}

		switch {
		case t.AssignableTo(timeType):
			v = v.MethodByName("String")

			buf.Writef("%q", v.Call(nil)[0])
			return
		}

		switch t.Kind() {
		case reflect.Map:
			buf.Writef("%s{%s", t.String(), nl)
			buf.Indent(1)

			kvs := make([][]reflect.Value, 0)

			iter := v.MapRange()
			for iter.Next() {
				k := iter.Key()
				v := iter.Value()
				kvs = append(kvs, []reflect.Value{k, v})
			}

			sort.Slice(kvs, func(i, j int) bool {
				k1 := fmt.Sprintf("%v", kvs[i][0].Interface())
				k2 := fmt.Sprintf("%v", kvs[j][0].Interface())

				return k1 < k2
			})

			for _, kv := range kvs {
				k, v := kv[0], kv[1]

				buf.WriteIndent()
				inner(k, false)
				buf.Writef(":")
				inner(v, false)
				if nl != "" {
					buf.Writef(",%s", nl)
				}
			}
			buf.Indent(-1)
			buf.WriteIndent()
			buf.Writef("}")
		case reflect.Slice:
			t = t.Elem()

			buf.Writef("[]%s{%s", t.String(), nl)
			buf.Indent(1)
			for i := 0; i < v.Len(); i++ {
				buf.WriteIndent()
				inner(v.Index(i), false)

				buf.Writef(",%s", nl)
			}
			buf.Indent(-1)
			buf.WriteIndent()
			buf.Writef("}")
		case reflect.Struct:
			if !slice {
				if isPtr {
					buf.Writef("&")
				}

				buf.Writef("%s{%s", t.String(), nl)
				buf.Indent(1)
			}

			for i := 0; i < v.NumField(); i++ {
				vf := v.Field(i)
				if vf.IsZero() {
					continue
				}

				tf := t.Field(i)

				if !unicode.IsUpper(rune(tf.Name[0])) {
					continue
				}

				buf.WriteIndent()
				buf.Writef("%s:", tf.Name)

				if buf.indent > 0 {
					buf.Writef(" %s", justify(t, i))
				} else {
					buf.Writef(" ")
				}

				inner(vf, false)

				if nl != "" || i < v.NumField()-1 {
					if nl == "" {
						buf.Writef(", ")
					} else {
						buf.Writef(",%s", nl)
					}
				}
			}

			if buf.indent > 0 {
				buf.Indent(-1)
			}

			if !slice {
				buf.WriteIndent()
				buf.Writef("}")
			}
		default:
			if xs := strings.Split(t.String(), "."); len(xs) > 1 {
				buf.Writef("%s.", xs[0])
			}

			switch x := v.Interface().(type) {
			case string:
				buf.Writef("%q", x)
			default:
				buf.Writef("%v", x)
			}
		}
	}

	v := reflect.ValueOf(value)

	if v.Type().Kind() == reflect.Ptr && v.IsZero() {
		buf.Writef("nil")
		return nil
	}

	inner(v, false)

	return nil
}

var timeType = reflect.TypeOf(time.Time{})

type writer struct {
	io.Writer
	indent int
}

func (w *writer) Indent(n int) {
	if w.indent >= 0 {
		w.indent += n
	}
}

func (w *writer) Writef(format string, args ...interface{}) error {
	_, err := w.Write([]byte(fmt.Sprintf(format, args...)))

	return err
}

func (w *writer) Writeln(s string) error {
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}

	if w.indent >= 0 {
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return nil
}

func (w *writer) WriteIndent() error {
	for i := 0; i < w.indent; i++ {
		if _, err := w.Write([]byte("\t")); err != nil {
			return err
		}
	}

	return nil
}

func justify(t reflect.Type, n int) string {
	var max int

	f := t.Field(n)

	switch f.Type.Kind() {
	case reflect.Ptr:
	case reflect.Struct:
	case reflect.Slice:
	default:
		if n := len(f.Name); n > max {
			max = n
		}
	}

	if max == 0 {
		return ""
	}

	for i := n - 1; i >= 0; i-- {
		f := t.Field(i)

		var ok bool

		switch f.Type.Kind() {
		case reflect.Ptr:
		case reflect.Struct:
		case reflect.Slice:
		default:
			ok = true

			if n := len(f.Name); n > max {
				max = n
			}
		}

		if !ok {
			break
		}
	}

	for i := n + 1; i < t.NumField(); i++ {
		f := t.Field(i)

		var ok bool

		switch f.Type.Kind() {
		case reflect.Ptr:
		case reflect.Struct:
		case reflect.Slice:
		default:
			ok = true

			if n := len(f.Name); n > max {
				max = n
			}
		}

		if !ok {
			break
		}
	}

	return strings.Repeat(" ", max-len(f.Name))
}
