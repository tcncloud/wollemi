package cobra

import (
	"github.com/spf13/cobra"

	"github.com/tcncloud/wollemi/ports/ctl"
)

func SymlinkGoPathCmd(app ctl.Application) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "go-path [path...]",
		Short: "symlink third party dependencies into go path",
		Long: Description(`
			Symlinks third party dependencies into the go path. Symlinks will not be created
			when there are existing files in the symlink path. Instead the command will
			issue warnings where symlink creation was not possible. When deletion of these
			files is acceptable the --force flag can be used to force symlink creation by
			first removing the existing files preventing the symlink creation.
		`),
		Example: Long(`
			Symlink all imported third party deps under the routes directory into the go path.
			    $ wollemi symlink go-path project/service/routes/...

			Symlink all third party deps for specific package into the go path.
			    $ wollemi symlink go-path project/service/routes
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			wollemi, err := app.Wollemi()
			if err != nil {
				return err
			}

			return wollemi.SymlinkGoPath(force, args)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "force symlink by removing existing files in symlink path")

	return cmd
}
