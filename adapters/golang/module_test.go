package golang_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tcncloud/wollemi/adapters/golang"
)

func TestModule_GoInfo_SinglePackage(t *testing.T) {
	pkg, err := golang.GoInfo("go/build")
	require.Nil(t, err)
	require.NotEmpty(t, pkg)
	require.Len(t, pkg, 1)
}

func TestModule_GoInfo_MultiplePackages(t *testing.T) {
	pkg, err := golang.GoInfo("go/...")
	require.Nil(t, err)
	require.NotEmpty(t, pkg)
	// NOTE: adapt the number of packages w/ golang version
	require.Len(t, pkg, 14)
}
