package cobra

import (
	"github.com/spf13/cobra"
)

func RulesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rules",
		Short: "build rule listing and deletion",
	}
}
