package cobra

import (
	"github.com/spf13/cobra"
	"github.com/tcncloud/wollemi/ports/ctl"
)

func GenerateCmd(app ctl.Application) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "generate BUILD files on an existing go.mod project",
		Long:  Description(``),
		RunE: func(cmd *cobra.Command, args []string) error {
			wollemi, err := app.Wollemi()
			if err != nil {
				return err
			}
			return wollemi.Generate(args)
		},
	}
	return cmd
}
