package cobra

import (
	"os"

	"github.com/spf13/cobra"
)

func CompletionBashCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "bash auto-completion",
		Long: Description(`
			The wollemi completion script for Bash can be generated with the command wollemi
			completion bash. Sourcing the completion script in your shell enables wollemi
			command auto-completion.

			To do so in all your shell sessions, add the following to your ~/.bash_profile file:

			source <(wollemi completion bash)

			After reloading your shell, wollemi autocompletion should be working.

			Wollemi completions require bash version 4.1 or higher and bash-completion@2. You
			can check your version by running echo $BASH_VERSION. If your version is too old
			and you are using macOS you can install or upgrade it using Homebrew.

			brew install bash
			brew install bash-completion@2
		`),
		Run: func(cmd *cobra.Command, args []string) {
			root.GenBashCompletion(os.Stdout)
		},
	}
}
