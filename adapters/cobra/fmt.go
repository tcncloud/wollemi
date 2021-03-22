package cobra

import (
	"github.com/spf13/cobra"

	"github.com/tcncloud/wollemi/ports/ctl"
	"github.com/tcncloud/wollemi/ports/wollemi"
)

func FmtCmd(app ctl.Application) *cobra.Command {
	config := wollemi.Config{
		Gofmt: wollemi.Gofmt{
			Rewrite: wollemi.Bool(false),
		},
	}

	cmd := &cobra.Command{
		Use:   "fmt [path...]",
		Short: "format build files",
		Long: Description(`
			Formats please build files. Formatting modifications include:
			  - Double quoted strings instead of single quoted strings.
			  - Deduplication of attribute list entries.
			  - Ordering of attribute list entries.
			  - Ordering of rule attributes.
			  - Deletion of empty build files.
			  - Consistent build identifiers.
			  - Text alignment.
		`),
		Example: Long(`
			Format a specific build file.
			    $ wollemi fmt project/service/routes

			Recursively format all build files under the routes directory.
			    $ wollemi fmt project/service/routes/...

			Recursively format all build files under the working directory.
			    $ wollemi fmt
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			wollemi, err := app.Wollemi()
			if err != nil {
				return err
			}

			return wollemi.Format(config, args)
		},
	}

	return cmd
}
