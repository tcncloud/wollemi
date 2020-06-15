package cobra

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tcncloud/wollemi/ports/ctl"
	"github.com/tcncloud/wollemi/ports/logging"
)

func RootCmd(app ctl.Application) *cobra.Command {
	var (
		logLevel  string
		logFormat string
	)

	cmd := &cobra.Command{
		Use:   "wollemi",
		Short: "cli for wollemi",
		Long: Long(`
			Please build file generator and formatter.
		`),
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if level, err := logging.ParseLevel(logLevel); err != nil {
				return fmt.Errorf("could not parse log level %q: %v", logLevel, err)
			} else {
				app.Logger().SetLevel(level)
			}

			if format, err := logging.ParseFormat(logFormat); err != nil {
				return fmt.Errorf("could not parse log format %q: %v", logFormat, err)
			} else {
				app.Logger().SetFormatter(format)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&logLevel, "log", "info", "logging level")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "logging format (text, json)")

	return cmd
}
