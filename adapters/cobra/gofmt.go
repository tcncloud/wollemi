package cobra

import (
	"github.com/spf13/cobra"

	"github.com/tcncloud/wollemi/ports/ctl"
	"github.com/tcncloud/wollemi/ports/wollemi"
)

func GoFmtCmd(app ctl.Application) *cobra.Command {
	config := wollemi.Config{}

	rewrite := config.Gofmt.GetRewrite()
	create := config.Gofmt.GetCreate()
	manage := config.Gofmt.GetManage()

	cmd := &cobra.Command{
		Use:   "gofmt [path...]",
		Short: "format and generate build files from existing go code",
		Long: Description(`
			Rewrites and generates go_binary, go_library and go_test rules according to
			existing go code. It also applies all formatting modifications from the
			wollemi fmt command.

			Occasionally a go dependency will be able to be resolved to multiple go
			get rules and wollemi may choose the wrong target for your needs. These
			cases can be resolved using a config file which sets a known dependency
			mapping from the go package to the desired target.

			# project/service/routes/.wollemi.json
			{
			  "default_visibility": "//project/service/routes/...",
			  "known_dependency": {
			    "github.com/olivere/elastic": "//third_party/go/github.com/olivere/elastic:v7"
			  }
			}

			Config files can be placed in any directory. Every build file gets
			formatted using a config which is the result of merging together all
			config files discovered between the build file directory and the
			directory gofmt was invoked from.

			The config file can also define a default visibility. When wollemi gofmt
			is invoked recursively on a directory it will use a default visibility
			equal to the path it was given on any new go build rules generated. The
			visibility of existing go build rules is never modified.

			For example, the following gofmt would apply a default visibility of
			["project/service/routes/..."] to any new go build rules generated.

			wollemi gofmt project/service/routes/...

			When gofmt is run on an individual package the default visibility applied is
			["PUBLIC"] for any new go build rules generated.

			Alternatively the default visibility can be explicitly provided through
			a .wollemi.json config file which will override both implicit cases above.

			Sometimes a third party dependency is required even though the go code
			doesn't directly require it. To force gofmt to keep these dependencies you
			must decorate the dependency with the following comment.

			"//third_party/go/cloud.google.com/go:container",  # wollemi:keep

			The keep comment can also be placed above go build rules you don't want gofmt
			to modify. These cases should be rare and this feature should be used only when
			absolutely necessary.
		`),
		Example: Long(`
			Go format a specific build file.
			    $ wollemi gofmt project/service/routes

			Recursively go format all build files under the routes directory.
			    $ wollemi gofmt project/service/routes/...

			Recursively go format all build files under the working directory.
			    $ wollemi gofmt
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			wollemi, err := app.Wollemi()
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("rewrite") {
				config.Gofmt.Rewrite = &rewrite
			}

			if cmd.Flags().Changed("create") {
				config.Gofmt.Create = create
			}

			if cmd.Flags().Changed("manage") {
				config.Gofmt.Manage = manage
			}

			return wollemi.GoFormat(config, args)
		},
	}

	cmd.Flags().BoolVar(&rewrite, "rewrite", rewrite, "allow rewriting of build files")
	cmd.Flags().StringSliceVar(&create, "create", create, "allow missing rule kinds to be created")
	cmd.Flags().StringSliceVar(&manage, "manage", manage, "allow existing rule kinds to be managed")

	return cmd
}
