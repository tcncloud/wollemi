package optional_test

import (
	"testing"

	"github.com/tcncloud/wollemi/domain/optional"

	"github.com/stretchr/testify/require"
)

func TestBool_IsTrue(t *testing.T) {
	for _, tt := range []struct {
		Name string
		Bool *optional.Bool
		Want bool
	}{{
		Name: "returns false when receiver unset",
	}, {
		Name: "returns false when receiver set to false",
		Bool: optional.BoolValue(false),
	}, {
		Name: "returns true when receiver set to true",
		Bool: optional.BoolValue(true),
		Want: true,
	}} {
		t.Run(tt.Name, func(t *testing.T) {
			require.Equal(t, tt.Want, tt.Bool.IsTrue())
		})
	}
}
