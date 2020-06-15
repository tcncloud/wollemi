package cobra

import (
	"github.com/spf13/cobra"

	"github.com/tcncloud/wollemi/ports/ctl"
)

func RulesUnusedCmd(app ctl.Application) *cobra.Command {
	var (
		prune   bool
		exclude []string
		kinds   []string
	)

	cmd := &cobra.Command{
		Use:   "unused [path...]",
		Short: "lists potentially unused build rules",
		Long: Description(`
			Lists potentially unused build rules. Unused in this context simply means no
			other build files depend on this rule. User discretion is needed to make the
			final call whether an unused build rule listed here should be pruned.
		`),
		Example: Long(`
			List all unused go_get rules.
			    $ wollemi rules unused --kind go_get

			List all unused rules under the routes directory.
			    $ wollemi rules unused project/service/routes/...

			List all unused rules except those under k8s and third_party.
			    $ wollemi rules unused --exclude k8s,third_party

			Prune unused third_party go_get rules.
			    $ wollemi rules unused --prune --kind go_get third_party/go/...
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			wollemi, err := app.Wollemi()
			if err != nil {
				return err
			}

			return wollemi.RulesUnused(prune, kinds, args, exclude)
		},
	}

	cmd.Flags().BoolVar(&prune, "prune", false, "prune matched rules")
	cmd.Flags().StringSliceVar(&kinds, "kind", kinds, "rule kinds to include (comma separated)")
	cmd.Flags().StringSliceVar(&exclude, "exclude", nil, "path prefixes to exclude (comma separated)")

	return cmd
}
