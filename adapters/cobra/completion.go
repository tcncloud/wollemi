package cobra

import (
	"github.com/spf13/cobra"
)

func CompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "shell auto-completion",
	}
}
