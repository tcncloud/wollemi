package expect

import (
	"bytes"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"

	"github.com/tcncloud/wollemi/domain/stringify"
)

type TestingT interface {
	Errorf(format string, args ...interface{})
	Helper()
}

func Equal(t TestingT, want, have interface{}) bool {
	if assert.ObjectsAreEqual(want, have) {
		return true
	}

	t.Helper()

	a := difflib.SplitLines(stringify.String(want, 0))
	b := difflib.SplitLines(stringify.String(have, 0))

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        a,
		B:        b,
		FromFile: "Want",
		ToFile:   "Have",
		Context:  len(a),
	})

	var buf bytes.Buffer

	buf.WriteRune('\n')

	for _, line := range difflib.SplitLines(diff) {
		if len(line) <= 0 {
			continue
		}

		switch line[0] {
		case '-':
			buf.WriteString("\033[32m")
		case '+':
			buf.WriteString("\033[31m")
		default:
			buf.WriteString("\033[0m")
		}

		buf.WriteString(line)
	}

	t.Errorf(buf.String())

	return false
}
