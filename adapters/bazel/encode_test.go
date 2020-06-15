package bazel_test

import (
	"testing"

	"github.com/bazelbuild/buildtools/build"
	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/adapters/bazel"
	"github.com/tcncloud/wollemi/ports/please"
)

var encode bazel.Encode

func TestEncode_StringExpr(t *testing.T) {
	have := factory.build.StringExpr("foo")
	want := factory.please.StringExpr("foo")

	require.Equal(t, want, encode.StringExpr(have))
	require.Equal(t, (*please.StringExpr)(nil), encode.StringExpr(nil))
}

func TestEncode_Ident(t *testing.T) {
	have := factory.build.Ident("True")
	want := factory.please.Ident("True")

	require.Equal(t, want, encode.Ident(have))
	require.Equal(t, (*please.Ident)(nil), encode.Ident(nil))
}

func TestEncode_CallExpr(t *testing.T) {
	have := factory.build.CallExpr(
		"go_library",
		factory.build.AssignExpr("name", "=", factory.build.StringExpr("api")),
		factory.build.AssignExpr("deps", "=", factory.build.ListExpr(
			"//third_party/go/github.com/spf13:viper",
			"//third_party/go/github.com/spf13:cobra",
			"//third_party/go/github.com/spf13:pflag",
		)),
	)

	want := factory.please.CallExpr(
		"go_library",
		factory.please.AssignExpr("name", "=", factory.please.StringExpr("api")),
		factory.please.AssignExpr("deps", "=", factory.please.ListExpr(
			"//third_party/go/github.com/spf13:viper",
			"//third_party/go/github.com/spf13:cobra",
			"//third_party/go/github.com/spf13:pflag",
		)),
	)

	require.Equal(t, want, encode.CallExpr(have))
	require.Equal(t, (*please.CallExpr)(nil), encode.CallExpr(nil))
}

func TestEncode_ListExpr(t *testing.T) {
	have := factory.build.ListExpr(
		"//third_party/go/github.com/spf13:viper",
		"//third_party/go/github.com/spf13:cobra",
		"//third_party/go/github.com/spf13:pflag",
	)

	want := factory.please.ListExpr(
		"//third_party/go/github.com/spf13:viper",
		"//third_party/go/github.com/spf13:cobra",
		"//third_party/go/github.com/spf13:pflag",
	)

	require.Equal(t, want, encode.ListExpr(have))
	require.Equal(t, (*please.ListExpr)(nil), encode.ListExpr(nil))
}

func TestEncode_BinaryExpr(t *testing.T) {
	have := factory.build.BinaryExpr(
		factory.build.StringExpr("foo"),
		"+",
		factory.build.StringExpr("bar"),
	)
	want := factory.please.BinaryExpr(
		factory.please.StringExpr("foo"),
		"+",
		factory.please.StringExpr("bar"),
	)

	require.Equal(t, want, encode.BinaryExpr(have))
	require.Equal(t, (*please.BinaryExpr)(nil), encode.BinaryExpr(nil))
}

func TestEncode_DictExpr(t *testing.T) {
	have := factory.build.DictExpr(
		factory.build.KeyValueExpr("one", "1"),
		factory.build.KeyValueExpr("two", "2"),
	)

	want := factory.please.DictExpr(
		factory.please.KeyValueExpr("one", "1"),
		factory.please.KeyValueExpr("two", "2"),
	)

	require.Equal(t, want, encode.DictExpr(have))
	require.Equal(t, (*please.DictExpr)(nil), encode.DictExpr(nil))
}

func TestEncode_KeyValueExpr(t *testing.T) {
	have := factory.build.KeyValueExpr("one", "1")
	want := factory.please.KeyValueExpr("one", "1")

	require.Equal(t, want, encode.KeyValueExpr(have))
	require.Equal(t, (*please.KeyValueExpr)(nil), encode.KeyValueExpr(nil))
}

func TestEncode_LiteralExpr(t *testing.T) {
	have := factory.build.LiteralExpr("1 + 1")
	want := factory.please.LiteralExpr("1 + 1")

	require.Equal(t, want, encode.LiteralExpr(have))
	require.Equal(t, (*please.LiteralExpr)(nil), encode.LiteralExpr(nil))
}

func TestEncode_AssignExpr(t *testing.T) {
	have := factory.build.AssignExpr("name", "=", factory.build.StringExpr("api"))
	want := factory.please.AssignExpr("name", "=", factory.please.StringExpr("api"))

	require.Equal(t, want, encode.AssignExpr(have))
	require.Equal(t, (*please.AssignExpr)(nil), encode.AssignExpr(nil))
}

func TestEncode_Comments(t *testing.T) {
	have := factory.build.Comments()
	want := factory.please.Comments()

	require.Equal(t, want, encode.Comments(have))
	require.Equal(t, please.Comments{}, encode.Comments(build.Comments{}))
}

func TestEncode_Expr(t *testing.T) {
	for _, tt := range []struct {
		Name string
		Have build.Expr
		Want please.Expr
	}{{
		Name: "can encode string expr",
		Have: factory.build.StringExpr("foo"),
		Want: factory.please.StringExpr("foo"),
	}, {
		Name: "can encode ident",
		Have: factory.build.Ident("True"),
		Want: factory.please.Ident("True"),
	}, {
		Name: "can encode call expr",
		Have: factory.build.CallExpr(
			"go_library",
			factory.build.AssignExpr("name", "=", factory.build.StringExpr("api")),
			factory.build.AssignExpr("deps", "=", factory.build.ListExpr(
				"//third_party/go/github.com/spf13:viper",
				"//third_party/go/github.com/spf13:cobra",
				"//third_party/go/github.com/spf13:pflag",
			)),
		),
		Want: factory.please.CallExpr(
			"go_library",
			factory.please.AssignExpr("name", "=", factory.please.StringExpr("api")),
			factory.please.AssignExpr("deps", "=", factory.please.ListExpr(
				"//third_party/go/github.com/spf13:viper",
				"//third_party/go/github.com/spf13:cobra",
				"//third_party/go/github.com/spf13:pflag",
			)),
		),
	}, {
		Name: "can encode list expr",
		Have: factory.build.ListExpr(
			"//third_party/go/github.com/spf13:viper",
			"//third_party/go/github.com/spf13:cobra",
			"//third_party/go/github.com/spf13:pflag",
		),
		Want: factory.please.ListExpr(
			"//third_party/go/github.com/spf13:viper",
			"//third_party/go/github.com/spf13:cobra",
			"//third_party/go/github.com/spf13:pflag",
		),
	}, {
		Name: "can encode assign expr",
		Have: factory.build.AssignExpr("name", "=", factory.build.StringExpr("api")),
		Want: factory.please.AssignExpr("name", "=", factory.please.StringExpr("api")),
	}, {
		Name: "can encode binary expr",
		Have: factory.build.BinaryExpr(
			factory.build.StringExpr("foo"),
			"+",
			factory.build.StringExpr("bar"),
		),
		Want: factory.please.BinaryExpr(
			factory.please.StringExpr("foo"),
			"+",
			factory.please.StringExpr("bar"),
		),
	}, {
		Name: "can encode dict expr",
		Have: factory.build.DictExpr(
			factory.build.KeyValueExpr("one", "1"),
			factory.build.KeyValueExpr("two", "2"),
		),
		Want: factory.please.DictExpr(
			factory.please.KeyValueExpr("one", "1"),
			factory.please.KeyValueExpr("two", "2"),
		),
	}, {
		Name: "can encode key value expr",
		Have: factory.build.KeyValueExpr("one", "1"),
		Want: factory.please.KeyValueExpr("one", "1"),
	}, {
		Name: "can encode literal expr",
		Have: factory.build.LiteralExpr("1 + 1"),
		Want: factory.please.LiteralExpr("1 + 1"),
	}} {
		t.Run(tt.Name, func(t *testing.T) {
			require.Equal(t, tt.Want, encode.Expr(tt.Have))
			require.Equal(t, nil, encode.Expr(nil))
		})
	}
}
