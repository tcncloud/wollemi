package golang_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/adapters/golang"
)

func TestImporter_ImportDir(t *testing.T) {
	if dir, ok := os.LookupEnv("PKG_DIR"); ok {
		require.NoError(t, os.MkdirAll(dir, os.FileMode(0700)))
		require.NoError(t, os.Chdir(dir))
	}

	t.Run("errors when package dir does not exist", func(t *testing.T) {
		importer := golang.NewImporter()

		pkg, err := importer.ImportDir("foo/bar", []string{"baz.go"})
		require.EqualError(t, err, "open foo/bar/baz.go: no such file or directory")
		require.Nil(t, pkg)
	})

	t.Run("gets package info for import directory", func(t *testing.T) {
		tmp, err := ioutil.TempDir("", "importer_test")
		require.NoError(t, err)

		defer os.RemoveAll(tmp)

		for _, x := range []struct {
			Path string
			Data []byte
		}{{
			Path: "adder/adder.go",
			Data: []byte(adder),
		}, {
			Path: "adder/adder_test.go",
			Data: []byte(adderTest),
		}, {
			Path: "multiplier/multiplier.go",
			Data: []byte(multiplier),
		}, {
			Path: "multiplier/multiplier_test.go",
			Data: []byte(multiplierTest),
		}} {
			path := filepath.Join(tmp, x.Path)

			require.NoError(t, os.MkdirAll(filepath.Dir(path), os.FileMode(0700)))
			require.NoError(t, ioutil.WriteFile(path, x.Data, os.FileMode(0644)))
		}

		importer := golang.NewImporter()

		// ---------------------------------------------------------------------

		have, err := importer.ImportDir(filepath.Join(tmp, "adder"), []string{
			"adder_test.go",
			"adder.go",
		})

		require.NoError(t, err)

		want := &golang.Package{
			Name: "adder",
			Imports: []string{
				"database/sql",
				"fmt",
				"github.com/spf13/viper",
			},
			XTestImports: []string{
				"github.com/stretchr/testify/require",
				"testing",
			},
			GoFiles: []string{
				"adder.go",
			},
			XTestGoFiles: []string{
				"adder_test.go",
			},
			GoFileImports: map[string][]string{
				"adder_test.go": []string{
					"testing",
					"github.com/stretchr/testify/require",
				},
				"adder.go": []string{
					"database/sql",
					"fmt",
					"github.com/spf13/viper",
				},
			},
		}

		require.Equal(t, want, have)

		// ---------------------------------------------------------------------

		have, err = importer.ImportDir(filepath.Join(tmp, "multiplier"), []string{
			"multiplier_test.go",
			"multiplier.go",
		})

		require.NoError(t, err)

		want = &golang.Package{
			Name: "multiplier",
			Imports: []string{
				"github.com/coreos/rkt/pkg/lock",
				"github.com/wollemi_test/project/proto",
				"github.com/wollemi_test/project/service/routes/async",
				"github.com/wollemi_test/project/service/routes/client",
				"github.com/wollemi_test/project/service/routes/server",
				"go/ast",
				"go/build",
			},
			TestImports: []string{
				"encoding/json",
				"fmt",
				"github.com/golang/mock/gomock",
				"github.com/stretchr/testify/require",
				"testing",
			},
			GoFiles: []string{
				"multiplier.go",
			},
			TestGoFiles: []string{
				"multiplier_test.go",
			},
			GoFileImports: map[string][]string{
				"multiplier_test.go": []string{
					"encoding/json",
					"fmt",
					"testing",
					"github.com/stretchr/testify/require",
					"github.com/golang/mock/gomock",
				},
				"multiplier.go": []string{
					"go/build",
					"go/ast",
					"github.com/coreos/rkt/pkg/lock",
					"github.com/wollemi_test/project/proto",
					"github.com/wollemi_test/project/service/routes/async",
					"github.com/wollemi_test/project/service/routes/client",
					"github.com/wollemi_test/project/service/routes/server",
				},
			},
		}

		require.Equal(t, want, have)
	})
}

const multiplier = `
package multiplier

import (
	"go/build"
	"go/ast"

	"github.com/coreos/rkt/pkg/lock"
	"github.com/wollemi_test/project/proto"
	"github.com/wollemi_test/project/service/routes/async"
	"github.com/wollemi_test/project/service/routes/client"
	"github.com/wollemi_test/project/service/routes/server"
)

func Multiply(x, y int) int {
	return x * y
}
`

const multiplierTest = `
package multiplier

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/golang/mock/gomock"
)

func TestMultiply(t*testing.T) {
	t.Run("multiples two integers", func(t*testing.T) {
		require.Equal(t, 14, multiply(7, 2))
	})
}
`

const adder = `
package adder

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
)

func NewAdder() *Adder {
	return &Adder{}
}

type Adder struct{}

func (*Adder) Add(x, y int) int {
	return x + y
}
`

const adderTest = `
package adder_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdder_Add(t *testing.T) {
	t.Run("adds two integers", func(t *testing.T) {
		res := adder.New().Add(1, 2)
		require.Equal(t, 3, res)
	})
}
`
