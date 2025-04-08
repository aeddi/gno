package convert

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/contribs/gasstat/pkg/balances"
	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newConvertBalancesCmd returns a command that allows users to convert balances file to another
// format. Supported formats are: json, txt and csv.
func newConvertBalancesCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "balances",
			ShortUsage: "convert balances <input-file> <output-file>",
			ShortHelp:  "convert balances file to another format",
			LongHelp:   "Convert balances file to another format. Supported formats are: json, txt and csv.\nThe format is determined by the file extension.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execConvertBalances(args, io)
		},
	)
}

func execConvertBalances(args []string, io commands.IO) error {
	// Get the input and output files from the command line arguments.
	inputFile, outputFile, err := file.GetInputOutputFiles(args, io, true)
	if err != nil {
		return err
	}

	// Get the balances from the input file.
	balances, err := balances.LoadFromFile(inputFile)
	if err != nil {
		return fmt.Errorf("error: failed to load balances from input file: %w", err)
	}

	// Save the balances to the output file.
	if err := balances.SaveToFile(outputFile); err != nil {
		return fmt.Errorf("error: failed to save balances to output file: %w", err)
	}

	io.Printfln("Balances converted from %s to %s", inputFile, outputFile)

	return nil
}
