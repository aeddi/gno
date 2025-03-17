package analyze

import "github.com/gnolang/gno/tm2/pkg/commands"

// NewAnalyzeCmd returns a command that allows users to process and analyze statistics
// from a data source (balances, usage or both).
func NewAnalyzeCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "analyze",
			ShortUsage: "<subcommand> <input-file>",
			ShortHelp:  "process and analyze statistics from a data source",
			LongHelp:   "Process and analyze statistics from a data source (balances, usage or both).",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newAnalyzeBalancesCmd(io),
	)

	return cmd
}
