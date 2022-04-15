package cobra

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tcncloud/wollemi/ports/ctl"
)

func SymlinkListCmd(app ctl.Application) *cobra.Command {
	var (
		broken  bool
		exclude []string
		gopath  bool
		name    string
		prune   bool
	)

	cmd := &cobra.Command{
		Use:   "list [path...]",
		Short: "list existing project symlinks",
		Long: Description(`
			Lists and optionally prunes project symlinks. Listed symlinks can be filtered
			with --broken in which case only broken symlinks are shown, --name in which
			case only symlinks with a name matching the provided pattern will be shown.
			For information on what the --name pattern can contain see go doc
			path/filepath.Match. Listed symlinks can also be filtered using --exclude
			in which case only symlinks which do not have the excluded prefix will be
			shown. When the --prune flag is added listed symlinks are deleted.
		`),
		Example: Long(`
			List all symlinks under the routes directory.
			    $ wollemi symlink list project/service/routes/...

			List only symlinks in a specific directory.
			    $ wollemi symlink list project/service/routes

			List all symlinks in the GOPATH. (excludes working directory)
			    $ wollemi symlink list --go-path

			List all broken symlinks.
			    $ wollemi symlink list --broken

			List all go_mock symlinks under the routes directory.
			    $ wollemi symlink list --name *.mg.go project/service/routes/...

			Prune all go_mock symlinks under the routes directory.
			    $ wollemi symlink list --prune --name *.mg.go project/service/routes/...
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			wollemi, err := app.Wollemi()
			if err != nil {
				return err
			}

			if gopath {
				if len(args) > 0 {
					return fmt.Errorf("cannot use --go-path with arguments")
				}

				exclude = append(exclude, wollemi.GoPkgPath())
				args = []string{filepath.Join(wollemi.GoSrcPath(), "...")}
			}

			if err := wollemi.SymlinkList(name, broken, prune, exclude, args); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&gopath, "go-path", false, "list symlinks which are in the go path")
	cmd.Flags().BoolVar(&broken, "broken", false, "list symlinks which are broken")
	cmd.Flags().BoolVar(&prune, "prune", false, "prune listed symlinks")
	cmd.Flags().StringVar(&name, "name", "*", "list symlinks matching name")
	cmd.Flags().StringSliceVar(&exclude, "exclude", nil, "list symlinks which exclude path prefix (comma separted)")

	return cmd
}
