package measure

import "github.com/gnolang/gno/tm2/pkg/commands"

// NewMeasureCmd returns a command that allows to measure gas usage for
// Realm function calls and bank transfers.
func NewMeasureCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "measure",
			ShortUsage: "<subcommand> <input-file>",
			ShortHelp:  "TODO",
			LongHelp:   "TODO",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands()

	return cmd
}
