package logrus_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/adapters/logrus"
	"github.com/tcncloud/wollemi/ports/logging"
)

func NewLoggerSuite(t *testing.T) *LoggerSuite {
	return &LoggerSuite{T: t}
}

type LoggerSuite struct {
	*testing.T
	stdout *bytes.Buffer
	log    *logrus.Logger
}

func (suite *LoggerSuite) It(name string, yield func(*LoggerSuite)) {
	suite.Helper()
	suite.Run(name, yield)
}

func (suite *LoggerSuite) Run(name string, yield func(*LoggerSuite)) {
	suite.Helper()
	suite.T.Run(name, func(t *testing.T) {
		suite := NewLoggerSuite(t)
		suite.Helper()

		suite.stdout = bytes.NewBuffer(nil)

		suite.log = logrus.NewLogger(suite.stdout)
		suite.log.SetFormatter(&logging.JsonFormatter{})

		yield(suite)
	})
}

type Entry struct {
	Level logging.Level
	Msg   string
	Time  time.Time
}

func (e *Entry) UnmarshalJSON(buf []byte) error {
	var m map[string]string

	err := json.Unmarshal(buf, &m)
	if err != nil {
		return err
	}

	if s, ok := m["level"]; ok {
		e.Level, err = logging.ParseLevel(s)
		if err != nil {
			return fmt.Errorf("could not parse level: %v", err)
		}
	}

	if s, ok := m["msg"]; ok {
		e.Msg = s
	}

	if s, ok := m["time"]; ok {
		e.Time, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return fmt.Errorf("could not parse time: %v", err)
		}
	}

	return nil
}

func (t *LoggerSuite) Decode(r io.Reader, entries ...interface{}) {
	decoder := json.NewDecoder(r)

	for _, entry := range entries {
		require.NoError(t, decoder.Decode(entry))
	}
}

func (t *LoggerSuite) NowUnix() time.Time {
	return time.Unix(time.Now().Unix(), 0)
}

func (t *LoggerSuite) BehavesLikeLogsMessage(run func(logging.Logger), lvl logging.Level, msg string) {
	type T = LoggerSuite

	t.Run(fmt.Sprintf("logs message with time at %s level", lvl), func(t *T) {
		t.log.SetLevel(lvl)

		min := t.NowUnix()
		run(t.log)
		max := t.NowUnix()

		delta := time.Duration(math.Ceil(float64(max.Sub(min)) / 2.0))

		entry := &Entry{}

		t.Decode(t.stdout, entry)
		assert.Equal(t, lvl, entry.Level)
		assert.Equal(t, msg, entry.Msg)
		assert.WithinDuration(t, min.Add(delta), entry.Time, delta)
	})

	t.Run(fmt.Sprintf("is not logged when log level less than %s", lvl), func(t *T) {
		t.log.SetLevel(lvl - 1)
		run(t.log)
		assert.Empty(t, t.stdout.String())
	})
}
