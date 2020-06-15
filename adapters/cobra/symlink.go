package cobra

import (
	"github.com/spf13/cobra"
)

func SymlinkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "symlink",
		Short: "symlink listing, creation and deletion",
	}
}
