package analyze

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/contribs/gasstat/pkg/distribution"
	"github.com/gnolang/gno/contribs/gasstat/pkg/stats"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newAnalyzeBalancesCmd returns a command that allows users process a balances file
// and produce statistics.
func newAnalyzeBalancesCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "balances",
			ShortUsage: "analyze balances <raw-balances-file>",
			ShortHelp:  "process a balances file to produce statistics",
			LongHelp:   "Process a balances file to produce statistics on distribution. Supported input formats are: json, txt and csv.\nThe format is determined by the file extension.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execAnalyzeBalances(args, io)
		},
	)
}

func execAnalyzeBalances(args []string, io commands.IO) error {
	// Check if the number of arguments is valid.
	if len(args) != 1 {
		io.ErrPrintln("error: invalid number of arguments, expected a balances input file")
		return flag.ErrHelp
	}

	var (
		inputFile  = args[0]
		outputFile = "./report.json" // TODO: add a flag for the output file.
	)

	// Get the balances from the input file.
	balances, err := distribution.LoadFromFile(inputFile)
	if err != nil {
		return fmt.Errorf("error: failed to load balances from input file: %w", err)
	}

	// Process the balances and produce a statistics report.
	statistics := stats.NewFromDistribution(balances)
	jsonBytes, err := json.MarshalIndent(statistics, "", "  ")
	if err != nil {
		return fmt.Errorf("error: failed to marshal statistics to JSON: %w", err)
	}

	// Write the statistics report to the output file.
	if err := os.WriteFile(outputFile, jsonBytes, 0644); err != nil {
		return fmt.Errorf("error: failed to write statistics to output file: %w", err)
	}

	io.Printfln("Statistics report written to %s", outputFile)

	return nil
}
