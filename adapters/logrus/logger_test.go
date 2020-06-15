package logrus_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tcncloud/wollemi/ports/logging"
)

func TestLogger_Infof(t *testing.T) {
	NewLoggerSuite(t).TestInfof()
}

func TestLogger_Warnf(t *testing.T) {
	NewLoggerSuite(t).TestWarnf()
}

func TestLogger_Warn(t *testing.T) {
	NewLoggerSuite(t).TestWarn()
}

func TestLogger_Debug(t *testing.T) {
	NewLoggerSuite(t).TestDebug()
}

func TestLogger_WithError(t *testing.T) {
	NewLoggerSuite(t).TestWithError()
}

func TestLogger_WithField(t *testing.T) {
	NewLoggerSuite(t).TestWithField()
}

func TestLogger_WithFields(t *testing.T) {
	NewLoggerSuite(t).TestWithFields()
}

func TestLogger_SetLevel(t *testing.T) {
	NewLoggerSuite(t).TestSetLevel()
}

// -----------------------------------------------------------------------------

func (t *LoggerSuite) TestInfof() {
	t.BehavesLikeLogsMessage(
		func(log logging.Logger) { log.Infof("hello: %q", "joe") },
		logging.InfoLevel,
		`hello: "joe"`,
	)
}

func (t *LoggerSuite) TestWarnf() {
	t.BehavesLikeLogsMessage(
		func(log logging.Logger) { log.Warnf("one: %d, two: %d", 1, 2) },
		logging.WarnLevel,
		`one: 1, two: 2`,
	)
}

func (t *LoggerSuite) TestWarn() {
	t.BehavesLikeLogsMessage(
		func(log logging.Logger) { log.Warn("winter is coming") },
		logging.WarnLevel,
		"winter is coming",
	)
}

func (t *LoggerSuite) TestDebug() {
	t.BehavesLikeLogsMessage(
		func(log logging.Logger) { log.Debug("windows") },
		logging.DebugLevel,
		"windows",
	)
}

func (t *LoggerSuite) TestWithError() {
	type T = LoggerSuite

	t.It("adds error to log entry", func(t *T) {
		t.log.WithError(fmt.Errorf("boom")).
			Infof("goes the dynamite")

		var entry struct {
			Error string `json:"error"`
		}

		t.Decode(t.stdout, &entry)
		assert.Equal(t, "boom", entry.Error)
	})
}

func (t *LoggerSuite) TestWithField() {
	type T = LoggerSuite

	t.It("adds field to log entry", func(t *T) {
		t.log.WithField("foo", "ab").
			WithField("bar", 1234).
			WithField("baz", true).
			Infof("")

		var entry struct {
			Foo string `json:"foo"`
			Bar int    `json:"bar"`
			Baz bool   `json:"baz"`
		}

		t.Decode(t.stdout, &entry)
		assert.Equal(t, "ab", entry.Foo)
		assert.Equal(t, 1234, entry.Bar)
		assert.Equal(t, true, entry.Baz)
	})
}

func (t *LoggerSuite) TestWithFields() {
	type T = LoggerSuite

	t.It("adds field to log entry", func(t *T) {
		t.log.WithFields(logging.Fields{
			"foo": "ab",
			"bar": 1234,
			"baz": true,
		}).Infof("")

		var entry struct {
			Foo string `json:"foo"`
			Bar int    `json:"bar"`
			Baz bool   `json:"baz"`
		}

		t.Decode(t.stdout, &entry)
		assert.Equal(t, "ab", entry.Foo)
		assert.Equal(t, 1234, entry.Bar)
		assert.Equal(t, true, entry.Baz)
	})
}

func (t *LoggerSuite) TestSetLevel() {
	type T = LoggerSuite

	for _, lvl := range []logging.Level{
		logging.PanicLevel,
		logging.FatalLevel,
		logging.ErrorLevel,
		logging.WarnLevel,
		logging.InfoLevel,
		logging.DebugLevel,
		logging.TraceLevel,
	} {
		t.It(fmt.Sprintf("sets log level to %s", lvl), func(t *T) {
			t.log.SetLevel(lvl)
			assert.Equal(t, lvl, t.log.GetLevel())
		})
	}
}
