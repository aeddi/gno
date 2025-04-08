package analyze

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/contribs/gasstat/pkg/gas"
	"github.com/gnolang/gno/contribs/gasstat/pkg/stats"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"gopkg.in/yaml.v3"
)

// newAnalyzeUsageCmd returns a command that allows users process one or more
// usage files and produce an usage statistics report.
func newAnalyzeUsageCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "usage",
			ShortUsage: "analyze usage <measurements-file-1> [<measurements-file-2> ...] <report-file>",
			ShortHelp:  "process one or more measurements files to produce usage statistics",
			LongHelp:   "Process one or more measurements file to produce statistics on usage. Supported measurements file formats are: json, txt and csv.\nSupported report file formats are: json, yaml and md.\nThe format is determined by the file extension.",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execAnalyzeUsage(args, io)
		},
	)
}

func execAnalyzeUsage(args []string, io commands.IO) error {
	// Check if the number of arguments is valid.
	if len(args) < 2 {
		io.ErrPrintln("error: invalid number of arguments, at least one measurements file and a report file are required")
		return flag.ErrHelp
	}

	var (
		err               error
		measurementsFiles = args[:len(args)-1]
		reportFile        = args[len(args)-1]
		measurements      = make([]*gas.Measurements, len(measurementsFiles))
	)

	// Check if the report file has a valid extension.
	reportFileFormat := file.FileFormatFromExt(reportFile)
	if reportFileFormat != file.JSON &&
		reportFileFormat != file.YAML &&
		reportFileFormat != file.MARKDOWN {
		io.ErrPrintln("error: invalid report file format, supported formats are: json, yaml and md")
		return flag.ErrHelp
	}

	// Get the measurements from the measurements files.
	for i := range measurementsFiles {
		measurements[i], err = gas.LoadFromFile(measurementsFiles[i])
		if err != nil {
			return fmt.Errorf("error: failed to load measurements from input file: %w", err)
		}
	}

	// Process the measurements and produce an usage statistics report.
	usage := stats.NewUsage(measurements)

	// Marshal the usage statistics to the specified format.
	var reportBytes []byte
	switch reportFileFormat {
	case file.JSON:
		reportBytes, err = json.MarshalIndent(usage, "", "  ")
		if err != nil {
			return fmt.Errorf("error: failed to marshal an usage statistics to JSON: %w", err)
		}

	case file.YAML:
		reportBytes, err = yaml.Marshal(usage)
		if err != nil {
			return fmt.Errorf("error: failed to marshal an usage statistics to YAML: %w", err)
		}

	case file.MARKDOWN:
		// TODO: Implement markdown format support.
		return errors.New("error: markdown format is not supported yet")
	}

	// Write the usage statistics to the report file.
	if err := os.WriteFile(reportFile, reportBytes, 0644); err != nil {
		return fmt.Errorf("error: failed to write usage statistics to output file: %w", err)
	}

	io.Printfln("Usage statistics report written to %s", reportFile)

	return nil
}
