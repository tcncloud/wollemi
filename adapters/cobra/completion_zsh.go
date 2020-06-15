package cobra

import (
	"os"

	"github.com/spf13/cobra"
)

func CompletionZshCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "zsh auto-completion",
		Long: Description(`
			The wollemi completion script for Zsh can be generated with the command wollemi
			completion zsh. Sourcing the completion script in your shell enables wollemi
			command auto-completion.

			To do so in all your shell sessions, add the following to your ~/.zshrc file:

			autoload -Uz compinit && compinit -C
			source <(wollemi completion zsh)
			compdef _wollemi wollemi

			After reloading your shell, wollemi autocompletion should be working.
		`),
		Run: func(cmd *cobra.Command, args []string) {
			root.GenZshCompletion(os.Stdout)
		},
	}
}
