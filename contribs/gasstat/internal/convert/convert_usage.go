package convert

import (
	"context"
	"errors"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newConvertUsageCmd returns a command that allows users to convert usage file to another
// format. Supported formats are: json and csv.
func newConvertUsageCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "usage",
			ShortUsage: "convert usage <input-file> <output-file>",
			ShortHelp:  "convert usage file to another format",
			LongHelp:   "Convert usage file to another format. Supported formats are: json and csv.\nThe format is determined by the file extension.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execConvertBalances(args, io)
		},
	)
}

func execConvertUsage(args []string, io commands.IO) error {
	_ = args
	_ = io
	return errors.New("execConvertBalances not implemented")
}
