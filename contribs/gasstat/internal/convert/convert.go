package convert

import (
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// NewConvertCmd returns a command that allows users to convert data sources to another format.
func NewConvertCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "convert",
			ShortUsage: "<subcommand> <input-file> <output-file>",
			ShortHelp:  "convert a data source to another format",
			LongHelp:   "Convert a data source (balances or usage) to another format. The format is determined by the file extension.",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newConvertBalancesCmd(io),
		newConvertUsageCmd(io),
	)

	return cmd
}
