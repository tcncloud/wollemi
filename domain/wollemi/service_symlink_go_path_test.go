package wollemi_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_SymlinkGoPath(t *testing.T) {
	NewServiceSuite(t).TestService_SymlinkGoPath()
}

func (t *ServiceSuite) TestService_SymlinkGoPath() {
	type T = ServiceSuite

	t.It("can symlink third party dependencies into gopath", func(t *T) {
		have := []string{
			"//:wollemi",
			"///adapters/bazel:bazel",
			"//third_party/go/github.com/bazelbuild/buildtools:buildtools",
			"//third_party/go/github.com/bazelbuild/buildtools:_buildtools-buildtools#download",
			"///ports/please:please",
			"///adapters/filesystem:filesystem",
			"//third_party/go/github.com/sirupsen:logrus",
			"//third_party/go/github.com/konsorten:go-windows-terminal-sequences",
			"//third_party/go/github.com/konsorten:_go-windows-terminal-sequences-go-windows-terminal-sequences#download",
			"//third_party/go/golang.org/x:sys",
			"//third_party/go/golang.org/x:_sys-sys#download",
			"//third_party/go/github.com/sirupsen:_logrus-logrus#download",
			"///ports/filesystem:filesystem",
			"///adapters/golang:golang",
			"///ports/golang:golang",
			"///adapters/please:please",
			"///domain/wollemi:wollemi",
			"///ports/wollemi:wollemi",
		}

		gopkg := "github.com/wollemi_test"
		gosrc := "/go/src"

		goSrcPath := func(elems ...string) string {
			return filepath.Join(gosrc, filepath.Join(elems...))
		}

		plzSrcPath := func(s string) string {
			return goSrcPath(gopkg, "plz-out/gen/third_party/go", s, "src", s)
		}

		t.please.EXPECT().QueryDeps(any).
			DoAndReturn(func(paths ...string) (*bufio.Reader, error) {
				assert.Equal(t, []string{
					"project/proto:all",
					"project/service/routes/...",
				}, paths)

				r := strings.NewReader(strings.Join(have, "\n"))
				return bufio.NewReader(r), nil
			})

		t.filesystem.EXPECT().Walk(goSrcPath("github.com/sirupsen"), any)
		t.filesystem.EXPECT().Walk(goSrcPath("github.com/bazelbuild/buildtools"), any)
		t.filesystem.EXPECT().Walk(goSrcPath("github.com/konsorten"), any)
		t.filesystem.EXPECT().Walk(goSrcPath("golang.org/x"), any)

		t.filesystem.EXPECT().Walk(plzSrcPath("github.com/sirupsen"), any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				name := "logrus"
				return ignoreSkipDir(walkFn(
					filepath.Join(plzSrcPath("github.com/sirupsen"), name),
					&FileInfo{FileName: name, FileIsDir: true},
					nil,
				))
			})

		t.filesystem.EXPECT().Walk(plzSrcPath("github.com/bazelbuild/buildtools"), any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				name := "build"
				return ignoreSkipDir(walkFn(
					filepath.Join(plzSrcPath("github.com/bazelbuild/buildtools"), name),
					&FileInfo{FileName: name, FileIsDir: true},
					nil,
				))
			})

		t.filesystem.EXPECT().Walk(plzSrcPath("github.com/konsorten"), any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				name := "go-windows-terminal-sequences"
				return ignoreSkipDir(walkFn(
					filepath.Join(plzSrcPath("github.com/konsorten"), name),
					&FileInfo{FileName: name, FileIsDir: true},
					nil,
				))
			})

		t.filesystem.EXPECT().Walk(plzSrcPath("golang.org/x"), any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				for _, name := range []string{"sync", "net", "text"} {
					err := ignoreSkipDir(walkFn(
						filepath.Join(plzSrcPath("golang.org/x"), name),
						&FileInfo{FileName: name, FileIsDir: true},
						nil,
					))

					if err != nil {
						return err
					}
				}

				return nil
			})

		t.filesystem.EXPECT().MkdirAll(goSrcPath("github.com/bazelbuild/buildtools"), os.FileMode(0744))
		t.filesystem.EXPECT().MkdirAll(goSrcPath("github.com/sirupsen"), os.FileMode(0744))
		t.filesystem.EXPECT().MkdirAll(goSrcPath("github.com/konsorten"), os.FileMode(0744))
		t.filesystem.EXPECT().MkdirAll(goSrcPath("golang.org/x"), os.FileMode(0744))

		t.filesystem.EXPECT().Readlink(filepath.Join(goSrcPath("github.com/konsorten"), "go-windows-terminal-sequences")).
			Return("", os.ErrNotExist)
		t.filesystem.EXPECT().Readlink(filepath.Join(goSrcPath("golang.org/x"), "sync")).
			Return("", os.ErrNotExist)
		t.filesystem.EXPECT().Readlink(filepath.Join(goSrcPath("golang.org/x"), "net")).
			Return("", os.ErrNotExist)
		t.filesystem.EXPECT().Readlink(filepath.Join(goSrcPath("golang.org/x"), "text")).
			Return("", os.ErrNotExist)
		t.filesystem.EXPECT().Readlink(filepath.Join(goSrcPath("github.com/bazelbuild/buildtools"), "build")).
			Return("", os.ErrNotExist)
		t.filesystem.EXPECT().Readlink(filepath.Join(goSrcPath("github.com/sirupsen"), "logrus")).
			Return("", os.ErrNotExist)

		t.filesystem.EXPECT().Symlink(
			filepath.Join(plzSrcPath("github.com/bazelbuild/buildtools"), "build"),
			filepath.Join(goSrcPath("github.com/bazelbuild/buildtools"), "build"),
		)
		t.filesystem.EXPECT().Symlink(
			filepath.Join(plzSrcPath("github.com/konsorten"), "go-windows-terminal-sequences"),
			filepath.Join(goSrcPath("github.com/konsorten"), "go-windows-terminal-sequences"),
		)
		t.filesystem.EXPECT().Symlink(
			filepath.Join(plzSrcPath("github.com/sirupsen"), "logrus"),
			filepath.Join(goSrcPath("github.com/sirupsen"), "logrus"),
		)
		t.filesystem.EXPECT().Symlink(
			filepath.Join(plzSrcPath("golang.org/x"), "sync"),
			filepath.Join(goSrcPath("golang.org/x"), "sync"),
		)
		t.filesystem.EXPECT().Symlink(
			filepath.Join(plzSrcPath("golang.org/x"), "net"),
			filepath.Join(goSrcPath("golang.org/x"), "net"),
		)
		t.filesystem.EXPECT().Symlink(
			filepath.Join(plzSrcPath("golang.org/x"), "text"),
			filepath.Join(goSrcPath("golang.org/x"), "text"),
		)

		t.DefaultMocks()

		wollemi := t.New(gosrc, gopkg)

		var (
			force = false
			paths = []string{
				"project/proto:all",
				"project/service/routes/...",
			}
		)

		wollemi.SymlinkGoPath(force, paths)
	})
}

func ignoreSkipDir(err error) error {
	if err == filepath.SkipDir {
		return nil
	}

	return err
}
