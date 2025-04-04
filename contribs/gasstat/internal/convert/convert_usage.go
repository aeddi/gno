package convert

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/contribs/gasstat/pkg/gas"
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
			return execConvertUsage(args, io)
		},
	)
}

func execConvertUsage(args []string, io commands.IO) error {
	// Get the input and output files from the command line arguments.
	inputFile, outputFile, err := file.GetInputOutputFiles(args, io, true)
	if err != nil {
		return err
	}

	// Get the gas usage from the input file.
	usage, err := gas.LoadFromFile(inputFile)
	if err != nil {
		return fmt.Errorf("error: failed to load gas usage from input file: %w", err)
	}

	// Save the gas usage to the output file.
	if err := usage.SaveToFile(outputFile); err != nil {
		return fmt.Errorf("error: failed to save gas usage to output file: %w", err)
	}

	io.Printfln("Gas usage converted from %s to %s", inputFile, outputFile)

	return nil
}
