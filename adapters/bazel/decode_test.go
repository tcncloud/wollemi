package bazel_test

import (
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/adapters/bazel"
	"github.com/tcncloud/wollemi/ports/please"
)

var decode bazel.Decode

func TestDecode_StringExpr(t *testing.T) {
	have := factory.please.StringExpr("foo")
	want := factory.build.StringExpr("foo")

	require.Equal(t, want, decode.StringExpr(have))
	require.Equal(t, (*build.StringExpr)(nil), decode.StringExpr(nil))
}

func TestDecode_Ident(t *testing.T) {
	have := factory.please.Ident("True")
	want := factory.build.Ident("True")

	require.Equal(t, want, decode.Ident(have))
	require.Equal(t, (*build.Ident)(nil), decode.Ident(nil))
}

func TestDecode_CallExpr(t *testing.T) {
	have := factory.please.CallExpr(
		"go_library",
		factory.please.AssignExpr("name", "=", factory.please.StringExpr("api")),
		factory.please.AssignExpr("deps", "=", factory.please.ListExpr(
			"//third_party/go/github.com/spf13:viper",
			"//third_party/go/github.com/spf13:cobra",
			"//third_party/go/github.com/spf13:pflag",
		)),
	)

	want := factory.build.CallExpr(
		"go_library",
		factory.build.AssignExpr("name", "=", factory.build.StringExpr("api")),
		factory.build.AssignExpr("deps", "=", factory.build.ListExpr(
			"//third_party/go/github.com/spf13:viper",
			"//third_party/go/github.com/spf13:cobra",
			"//third_party/go/github.com/spf13:pflag",
		)),
	)

	require.Equal(t, want, decode.CallExpr(have))
	require.Equal(t, (*build.CallExpr)(nil), decode.CallExpr(nil))
}

func TestDecode_ListExpr(t *testing.T) {
	have := factory.please.ListExpr(
		"//third_party/go/github.com/spf13:viper",
		"//third_party/go/github.com/spf13:cobra",
		"//third_party/go/github.com/spf13:pflag",
	)

	want := factory.build.ListExpr(
		"//third_party/go/github.com/spf13:viper",
		"//third_party/go/github.com/spf13:cobra",
		"//third_party/go/github.com/spf13:pflag",
	)

	require.Equal(t, want, decode.ListExpr(have))
	require.Equal(t, (*build.ListExpr)(nil), decode.ListExpr(nil))
}

func TestDecode_BinaryExpr(t *testing.T) {
	have := factory.please.BinaryExpr(
		factory.please.StringExpr("foo"),
		"+",
		factory.please.StringExpr("bar"),
	)

	want := factory.build.BinaryExpr(
		factory.build.StringExpr("foo"),
		"+",
		factory.build.StringExpr("bar"),
	)

	require.Equal(t, want, decode.BinaryExpr(have))
	require.Equal(t, (*build.BinaryExpr)(nil), decode.BinaryExpr(nil))
}

func TestDecode_DictExpr(t *testing.T) {
	have := factory.please.DictExpr(
		factory.please.KeyValueExpr("one", "1"),
		factory.please.KeyValueExpr("two", "2"),
	)

	want := factory.build.DictExpr(
		factory.build.KeyValueExpr("one", "1"),
		factory.build.KeyValueExpr("two", "2"),
	)

	require.Equal(t, want, decode.DictExpr(have))
	require.Equal(t, (*build.DictExpr)(nil), decode.DictExpr(nil))
}

func TestDecode_KeyValueExpr(t *testing.T) {
	have := factory.please.KeyValueExpr("one", "1")
	want := factory.build.KeyValueExpr("one", "1")

	require.Equal(t, want, decode.KeyValueExpr(have))
	require.Equal(t, (*build.KeyValueExpr)(nil), decode.KeyValueExpr(nil))
}

func TestDecode_LiteralExpr(t *testing.T) {
	have := factory.please.LiteralExpr("1 + 1")
	want := factory.build.LiteralExpr("1 + 1")

	require.Equal(t, want, decode.LiteralExpr(have))
	require.Equal(t, (*build.LiteralExpr)(nil), decode.LiteralExpr(nil))
}

func TestDecode_AssignExpr(t *testing.T) {
	have := factory.please.AssignExpr("name", "=", factory.please.StringExpr("api"))
	want := factory.build.AssignExpr("name", "=", factory.build.StringExpr("api"))

	require.Equal(t, want, decode.AssignExpr(have))
	require.Equal(t, (*build.AssignExpr)(nil), decode.AssignExpr(nil))
}

func TestDecode_Expr(t *testing.T) {
	for _, tt := range []struct {
		Name string
		Have please.Expr
		Want build.Expr
	}{{
		Name: "can encode string expr",
		Have: factory.please.StringExpr("foo"),
		Want: factory.build.StringExpr("foo"),
	}, {
		Name: "can encode ident",
		Have: factory.please.Ident("True"),
		Want: factory.build.Ident("True"),
	}, {
		Name: "can encode call expr",
		Have: factory.please.CallExpr(
			"go_library",
			factory.please.AssignExpr("name", "=", factory.please.StringExpr("api")),
			factory.please.AssignExpr("deps", "=", factory.please.ListExpr(
				"//third_party/go/github.com/spf13:viper",
				"//third_party/go/github.com/spf13:cobra",
				"//third_party/go/github.com/spf13:pflag",
			)),
		),
		Want: factory.build.CallExpr(
			"go_library",
			factory.build.AssignExpr("name", "=", factory.build.StringExpr("api")),
			factory.build.AssignExpr("deps", "=", factory.build.ListExpr(
				"//third_party/go/github.com/spf13:viper",
				"//third_party/go/github.com/spf13:cobra",
				"//third_party/go/github.com/spf13:pflag",
			)),
		),
	}, {
		Name: "can encode list expr",
		Have: factory.please.ListExpr(
			"//third_party/go/github.com/spf13:viper",
			"//third_party/go/github.com/spf13:cobra",
			"//third_party/go/github.com/spf13:pflag",
		),
		Want: factory.build.ListExpr(
			"//third_party/go/github.com/spf13:viper",
			"//third_party/go/github.com/spf13:cobra",
			"//third_party/go/github.com/spf13:pflag",
		),
	}, {
		Name: "can encode assign expr",
		Have: factory.please.AssignExpr("name", "=", factory.please.StringExpr("api")),
		Want: factory.build.AssignExpr("name", "=", factory.build.StringExpr("api")),
	}, {
		Name: "can encode binary expr",
		Have: factory.please.BinaryExpr(
			factory.please.StringExpr("foo"),
			"+",
			factory.please.StringExpr("bar"),
		),
		Want: factory.build.BinaryExpr(
			factory.build.StringExpr("foo"),
			"+",
			factory.build.StringExpr("bar"),
		),
	}, {
		Name: "can encode dict expr",
		Have: factory.please.DictExpr(
			factory.please.KeyValueExpr("one", "1"),
			factory.please.KeyValueExpr("two", "2"),
		),
		Want: factory.build.DictExpr(
			factory.build.KeyValueExpr("one", "1"),
			factory.build.KeyValueExpr("two", "2"),
		),
	}, {
		Name: "can encode key value expr",
		Have: factory.please.KeyValueExpr("one", "1"),
		Want: factory.build.KeyValueExpr("one", "1"),
	}, {
		Name: "can encode literal expr",
		Have: factory.please.LiteralExpr("1 + 1"),
		Want: factory.build.LiteralExpr("1 + 1"),
	}} {
		t.Run(tt.Name, func(t *testing.T) {
			require.Equal(t, tt.Want, decode.Expr(tt.Have))
			require.Equal(t, nil, decode.Expr(nil))
		})
	}
}
