package analyze

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/contribs/gasstat/pkg/balances"
	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/contribs/gasstat/pkg/stats"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"gopkg.in/yaml.v3"
)

// newAnalyzeDistributionCmd returns a command that allows users process a balances file
// and produce a distribution statistics report.
func newAnalyzeDistributionCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "distribution",
			ShortUsage: "analyze distribution <balances-file> <report-file>",
			ShortHelp:  "process a balances file to produce distribution statistics",
			LongHelp:   "Process a balances file to produce statistics on distribution. Supported balances file formats are: json, txt and csv.\nSupported report file formats are: json, yaml and md.\nThe format is determined by the file extension.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execAnalyzeDistribution(args, io)
		},
	)
}

func execAnalyzeDistribution(args []string, io commands.IO) error {
	// Get the balances and report files from the command line arguments.
	balancesFile, reportFile, err := file.GetInputOutputFiles(args, io, false)
	if err != nil {
		return err
	}

	// Check if the report file has a valid extension.
	reportFileFormat := file.FileFormatFromExt(reportFile)
	if reportFileFormat != file.JSON &&
		reportFileFormat != file.YAML &&
		reportFileFormat != file.MARKDOWN {
		io.ErrPrintln("error: invalid report file format, supported formats are: json, yaml and md")
		return flag.ErrHelp
	}

	// Get the balances from the input file.
	balances, err := balances.LoadFromFile(balancesFile)
	if err != nil {
		return fmt.Errorf("error: failed to load balances from input file: %w", err)
	}

	// Process the balances and produce a distribution distribution report.
	distribution := stats.NewDistribution(balances)

	// Marshal the distribution statistics to the specified format.
	var reportBytes []byte
	switch reportFileFormat {
	case file.JSON:
		reportBytes, err = json.MarshalIndent(distribution, "", "  ")
		if err != nil {
			return fmt.Errorf("error: failed to marshal distribution statistics to JSON: %w", err)
		}

	case file.YAML:
		reportBytes, err = yaml.Marshal(distribution)
		if err != nil {
			return fmt.Errorf("error: failed to marshal distribution statistics to YAML: %w", err)
		}

	case file.MARKDOWN:
		// TODO: Implement markdown format support.
		return errors.New("error: markdown format is not supported yet")
	}

	// Write the distribution statistics to the report file.
	if err := os.WriteFile(reportFile, reportBytes, 0644); err != nil {
		return fmt.Errorf("error: failed to write distribution statistics to output file: %w", err)
	}

	io.Printfln("Distribution statistics report written to %s", reportFile)

	return nil
}
