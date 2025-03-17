package convert

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/contribs/gasstat/pkg/distribution"
	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// NewConvertBalancesCmd returns a command that allows users to convert balances file to another
// format. Supported formats are: json, txt and csv.
func NewConvertBalancesCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "convert-balances",
			ShortUsage: "convert-balances <input-file> <output-file>",
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
	// Check if the number of arguments is valid.
	if len(args) != 2 {
		io.ErrPrintln("error: invalid number of arguments, expected an input file and an output file")
		return flag.ErrHelp
	}

	var (
		inputFile    = args[0]
		outputFile   = args[1]
		inputFormat  = file.FileFormatFromExt(inputFile)
		outputFormat = file.FileFormatFromExt(outputFile)
	)

	// Check if the input file format is supported.
	if inputFormat == file.Unknown {
		io.ErrPrintfln("error: unsupported input file format: %s", inputFile)
		return flag.ErrHelp
	}

	// Check if the output file format is supported.
	if outputFormat == file.Unknown {
		io.ErrPrintfln("error: unsupported output file format: %s", outputFile)
		return flag.ErrHelp
	}

	// Check if the input and output file formats are the same.
	if inputFormat == outputFormat {
		io.ErrPrintfln("error: input and output file formats must be different: %s", inputFormat)
		return flag.ErrHelp
	}

	// Get the balances from the input file.
	balances, err := distribution.LoadFromFile(inputFile)
	if err != nil {
		return fmt.Errorf("error: failed to load balances from input file: %w", err)
	}

	// Save the balances to the output file.
	if err := balances.SaveToFile(outputFile); err != nil {
		return fmt.Errorf("error: failed to save balances to output file: %w", err)
	}

	io.Printfln("balances converted from %s to %s", inputFile, outputFile)

	return nil
}
