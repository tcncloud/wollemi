package cobra

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tcncloud/wollemi/ports/ctl"
	"github.com/tcncloud/wollemi/ports/logging"
	"github.com/tcncloud/wollemi/ports/wollemi"
)

func RootCmd(app ctl.Application) *cobra.Command {
	var (
		logLevel  string
		logFormat string
	)

	cmd := &cobra.Command{
		Use:     "wollemi",
		Version: wollemi.WollemiVersion,
		Short:   "cli for wollemi",
		Long: Description(`
			Please build file generator and formatter capable of generating go_binary,
			go_library and go_test build rules from existing go code.
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
