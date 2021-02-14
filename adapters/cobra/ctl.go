package cobra

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/tcncloud/wollemi/ports/ctl"
)

func Ctl(app ctl.Application) *cobra.Command {
	var (
		fmt            = FmtCmd(app)
		gofmt          = GoFmtCmd(app)
		root           = RootCmd(app)
		symlink        = SymlinkCmd()
		symlinkGoPath  = SymlinkGoPathCmd(app)
		symlinkList    = SymlinkListCmd(app)
		rules          = RulesCmd()
		rulesUnused    = RulesUnusedCmd(app)
		completion     = CompletionCmd()
		completionBash = CompletionBashCmd(root)
		completionZsh  = CompletionZshCmd(root)
		generate       = GenerateCmd(app)
	)

	cmds := []*cobra.Command{
		fmt,
		gofmt,
		root,
		symlink,
		symlinkGoPath,
		symlinkList,
		rules,
		rulesUnused,
		completion,
		completionBash,
		completionZsh,
	}

	for _, cmd := range cmds {
		if !cmd.HasAvailableSubCommands() {
			for i := 0; i < 10; i++ {
				cmd.MarkZshCompPositionalArgumentFile(i)
			}
		}
	}

	addCommands(rules, rulesUnused)
	addCommands(symlink, symlinkGoPath, symlinkList)
	addCommands(completion, completionBash, completionZsh)
	addCommands(root, fmt, gofmt, symlink, rules, completion, generate)

	return root
}

func addCommands(cmd *cobra.Command, subCmds ...*cobra.Command) {
	for _, subCmd := range subCmds {
		cmd.AddCommand(subCmd)
	}

	setValidArgs(cmd, subCmds...)
}

func setValidArgs(cmd *cobra.Command, subCmds ...*cobra.Command) {
	cmd.ValidArgs = make([]string, len(subCmds))

	for i, subCmd := range subCmds {
		if n := strings.Index(subCmd.Use, " "); n > 0 {
			cmd.ValidArgs[i] = subCmd.Use[:n]
		} else {
			cmd.ValidArgs[i] = subCmd.Use
		}

	}
}

func Long(s string) string {
	var buf bytes.Buffer
	var prefix string
	var nl int

	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			nl++
			continue
		}

		if prefix == "" {
			if n := strings.IndexFunc(line, NotSpace); n > 0 {
				prefix = line[:n]
			} else {
				prefix = line
			}
		}

		if nl > 0 {
			if buf.Len() != 0 {
				buf.WriteString(strings.Repeat("\n", nl))
			}

			nl = 0
		}

		buf.WriteString("  ")
		buf.WriteString(strings.TrimPrefix(line, prefix))
		buf.WriteRune('\n')
	}

	return buf.String()
}

func Description(s string) string {
	return "Description:\n" + Long(s)
}

func NotSpace(r rune) bool {
	return !unicode.IsSpace(r)
}
