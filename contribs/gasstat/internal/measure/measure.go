package measure

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/contribs/gasstat/pkg/file"
	"github.com/gnolang/gno/contribs/gasstat/pkg/gas"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

const defaultRemote = "127.0.0.1:26657"

type measureFlags struct {
	remote string
}

// NewMeasureCmd returns a command that allows to measure gas usage for
// Realm function calls and bank transfers.
func NewMeasureCmd(io commands.IO) *commands.Command {
	var cfg measureFlags

	return commands.NewCommand(
		commands.Metadata{
			Name:       "measure",
			ShortUsage: "measure <config-file> <result-file>",
			ShortHelp:  "measure gas usage for Realm function calls and bank transfers",
			LongHelp:   "Measure gas usage for Realm function calls and bank transfers.\nSupported config formats are: json and yaml.\nSupported result formats are: json, txt and csv.\nThe format is determined by the file extension.",
		},
		&cfg,
		func(_ context.Context, args []string) error {
			return execMeasure(args, io, cfg)
		},
	)
}

func (mf *measureFlags) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&mf.remote,
		"remote",
		defaultRemote,
		"rpc address of the remote node to connect to",
	)
}

func execMeasure(args []string, io commands.IO, flags measureFlags) error {
	// Get the config and result files from the command line arguments.
	configFile, resultFile, err := file.GetInputOutputFiles(args, io, false)
	if err != nil {
		return err
	}

	// Measure gas usage using the config file.
	usage, err := gas.Measure(flags.remote, configFile)
	if err != nil {
		return fmt.Errorf("error: failed to measure gas usage: %w", err)
	}

	// Save the usage to the result file.
	if err := usage.SaveToFile(resultFile); err != nil {
		return fmt.Errorf("error: failed to save gas usage to result file: %w", err)
	}

	io.Printfln("Gas usage saved to %s", resultFile)

	return nil
}
