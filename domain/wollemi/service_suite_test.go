package wollemi_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/tcncloud/wollemi/domain/wollemi"
	"github.com/tcncloud/wollemi/ports/golang/mock"
	"github.com/tcncloud/wollemi/ports/please/mock"
	"github.com/tcncloud/wollemi/ports/wollemi/mock"
	"github.com/tcncloud/wollemi/testdata/mem"
)

var any = gomock.Any()

func NewServiceSuite(t *testing.T) *ServiceSuite {
	return &ServiceSuite{T: t}
}

type ServiceSuite struct {
	*testing.T
	ctrl       *gomock.Controller
	runner     *wollemi.Service
	logger     *mem.Logger
	filesystem *mock_wollemi.MockFilesystem
	golang     *mock_golang.MockImporter
	please     *mock_please.MockBuilder
}

func (suite *ServiceSuite) It(name string, yield func(*ServiceSuite)) {
	suite.Helper()
	suite.Run(name, yield)
}

func (suite *ServiceSuite) Run(name string, yield func(*ServiceSuite)) {
	suite.Helper()
	suite.T.Run(name, func(t *testing.T) {
		suite := NewServiceSuite(t)
		suite.Helper()

		suite.ctrl = gomock.NewController(&backgroundReporter{t: t})
		defer suite.ctrl.Finish()

		suite.logger = mem.NewLogger()
		suite.filesystem = mock_wollemi.NewMockFilesystem(suite.ctrl)
		suite.golang = mock_golang.NewMockImporter(suite.ctrl)
		suite.please = mock_please.NewMockBuilder(suite.ctrl)

		yield(suite)

		for _, line := range suite.logger.Lines() {
			t.Log(fmt.Sprint(line))
		}
	})
}

func (suite *ServiceSuite) New(gosrc, gopkg string) *wollemi.Service {
	return wollemi.New(
		suite.logger,
		suite.filesystem,
		suite.golang,
		suite.please,
		gosrc,
		gopkg,
	)
}

func (suite *ServiceSuite) DefaultMocks() {}

type backgroundReporter struct {
	t gomock.TestReporter

	mu     sync.Mutex
	fatalf []string
	errorf []string
}

func (br *backgroundReporter) Fatalf(msg string, args ...interface{}) {
	br.mu.Lock()
	defer br.mu.Unlock()
	s := fmt.Sprintf(msg, args...)
	br.fatalf = append(br.fatalf, s)
	panic(s)
}

func (br *backgroundReporter) Errorf(msg string, args ...interface{}) {
	br.mu.Lock()
	defer br.mu.Unlock()
	s := fmt.Sprintf(msg, args...)
	br.errorf = append(br.errorf, s)
	panic(s)
}

func (br *backgroundReporter) Finish() {
	br.mu.Lock()
	defer br.mu.Unlock()
	if len(br.errorf) != 0 {
		m := strings.Join(br.errorf, "\n  ")
		br.t.Errorf("delayed Errorf calls from a mock:\n  %s", m)
	}
	if len(br.fatalf) != 0 {
		m := strings.Join(br.fatalf, "\n  ")
		br.t.Fatalf("delayed Fatalf calls from a mock:\n  %s", m)
	}
}
