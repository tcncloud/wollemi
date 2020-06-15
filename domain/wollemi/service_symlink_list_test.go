package wollemi_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_SymlinkList(t *testing.T) {
	NewServiceSuite(t).TestService_SymlinkList()
}

func (t *ServiceSuite) TestService_SymlinkList() {
	type T = ServiceSuite

	const (
		gopkg = "github.com/wollemi_test"
		gosrc = "/go/src"
	)

	t.It("can list all project symlinks", func(t *T) {
		data := t.GoFormatTestData()

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					if err := walkFn(path, info, nil); err != nil {
						return err
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		wollemi := t.New(gosrc, gopkg)

		var (
			name    = "*"
			broken  bool
			prune   bool
			exclude []string
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/service.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/service.pb.go",
			},
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/entities.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/entities.pb.go",
			},
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/mock/mock.mg.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/mock/mock.mg.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})

	t.It("can list only broken symlinks", func(t *T) {
		data := t.GoFormatTestData()

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					if err := walkFn(path, info, nil); err != nil {
						return err
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		t.filesystem.EXPECT().Stat(any).AnyTimes().
			DoAndReturn(func(path string) (interface{}, error) {
				prefix := filepath.Join(gosrc, gopkg, "plz-out/gen") + "/"

				assert.Regexp(t, "^"+prefix, path)

				switch strings.TrimPrefix(path, prefix) {
				case "app/protos/entities.pb.go":
					return nil, os.ErrNotExist
				case "app/protos/mock/mock.mg.go":
					return nil, os.ErrNotExist
				}

				info := &FileInfo{
					FileName: filepath.Base(path),
					FileMode: os.FileMode(420),
				}

				return info, nil
			})

		wollemi := t.New(gosrc, gopkg)

		var (
			name         = "*"
			broken  bool = true
			prune   bool
			exclude []string
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/entities.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/entities.pb.go",
			},
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/mock/mock.mg.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/mock/mock.mg.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})

	t.It("can list symlinks matching name", func(t *T) {
		data := t.GoFormatTestData()

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					if err := walkFn(path, info, nil); err != nil {
						return err
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		wollemi := t.New(gosrc, gopkg)

		var (
			name    = "*.mg.go"
			broken  bool
			prune   bool
			exclude []string
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/mock/mock.mg.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/mock/mock.mg.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})

	t.It("can prune matched symlinks", func(t *T) {
		data := t.GoFormatTestData()

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					if err := walkFn(path, info, nil); err != nil {
						return err
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		t.filesystem.EXPECT().Stat(any).AnyTimes().
			DoAndReturn(func(path string) (interface{}, error) {
				prefix := filepath.Join(gosrc, gopkg, "plz-out/gen") + "/"

				assert.Regexp(t, "^"+prefix, path)

				switch strings.TrimPrefix(path, prefix) {
				case "app/protos/service.pb.go":
					return nil, os.ErrNotExist
				}

				info := &FileInfo{
					FileName: filepath.Base(path),
					FileMode: os.FileMode(420),
				}

				return info, nil
			})

		t.filesystem.EXPECT().Remove("app/protos/service.pb.go")

		wollemi := t.New(gosrc, gopkg)

		var (
			name         = "*.pb.go"
			broken  bool = true
			prune   bool = true
			exclude []string
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/service.pb.go",
				"msg":   "symlink deleted",
				"path":  "plz-out/gen/app/protos/service.pb.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})

	t.It("can exclude listing symlinks by path prefix", func(t *T) {
		data := t.GoFormatTestData()

		skipDir := make(map[string]struct{})

		t.filesystem.EXPECT().Walk(any, any).
			DoAndReturn(func(root string, walkFn filepath.WalkFunc) error {
				assert.Equal(t, ".", root)

			Walk:
				for _, path := range data.Walk {
					info := data.Lstat[path]
					if info == nil {
						continue
					}

					for prefix, _ := range skipDir {
						if strings.HasPrefix(path, prefix) {
							continue Walk
						}
					}

					err := walkFn(path, info, nil)
					if err == filepath.SkipDir {
						skipDir[path] = struct{}{}
					}
				}

				return nil
			})

		t.filesystem.EXPECT().Readlink(any).AnyTimes().
			DoAndReturn(func(path string) (string, error) {
				s, ok := data.Readlink[path]
				if !ok {
					t.Errorf("unexpected call to filesystem readlink: %s", path)
				}

				return filepath.Join(gosrc, gopkg, s), nil
			})

		wollemi := t.New(gosrc, gopkg)

		var (
			name    = "*"
			broken  bool
			prune   bool
			exclude []string = []string{
				"app/protos/mock",
			}
			include []string
		)

		wollemi.SymlinkList(name, broken, prune, exclude, include)

		want := []map[string]interface{}{
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/service.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/service.pb.go",
			},
			map[string]interface{}{
				"level": "info",
				"link":  "app/protos/entities.pb.go",
				"msg":   "symlink",
				"path":  "plz-out/gen/app/protos/entities.pb.go",
			},
		}

		for _, entry := range t.logger.Lines() {
			delete(entry, "time")
		}

		assert.ElementsMatch(t, want, t.logger.Lines())
	})
}
