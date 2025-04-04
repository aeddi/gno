package file

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// GetInputOutputFiles checks the input and output files for the convert subcommands.
func GetInputOutputFiles(args []string, io commands.IO, mustDiff bool) (string, string, error) {
	// Check if the number of arguments is valid.
	if len(args) != 2 {
		io.ErrPrintln("error: invalid number of arguments, expected an input file and an output file")
		return "", "", flag.ErrHelp
	}

	var (
		inputFile    = args[0]
		outputFile   = args[1]
		inputFormat  = FileFormatFromExt(inputFile)
		outputFormat = FileFormatFromExt(outputFile)
	)

	// Check if the input file format is supported.
	if inputFormat == Unknown {
		io.ErrPrintfln("error: unsupported input file format: %s", inputFile)
		return "", "", flag.ErrHelp
	}

	// Check if the output file format is supported.
	if outputFormat == Unknown {
		io.ErrPrintfln("error: unsupported output file format: %s", outputFile)
		return "", "", flag.ErrHelp
	}

	// Check if the input and output file formats are the same.
	if mustDiff && inputFormat == outputFormat {
		io.ErrPrintfln("error: input and output file formats must be different: %s", inputFormat)
		return "", "", flag.ErrHelp
	}

	return inputFile, outputFile, nil
}
