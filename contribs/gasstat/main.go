package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/contribs/gasstat/internal/analyze"
	"github.com/gnolang/gno/contribs/gasstat/internal/convert"
	"github.com/gnolang/gno/contribs/gasstat/internal/measure"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	io := commands.NewDefaultIO()

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Gas statistics collection and analysis tool",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		analyze.NewAnalyzeCmd(io),
		convert.NewConvertCmd(io),
		measure.NewMeasureCmd(io),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}
