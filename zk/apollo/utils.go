package apollo

import (
	"flag"

	"github.com/urfave/cli/v2"
)

func createMockContext(flags []cli.Flag) *cli.Context {
	set := flag.NewFlagSet("", flag.ContinueOnError)
	for _, f := range flags {
		f.Apply(set)
	}

	context := cli.NewContext(nil, set, nil)
	return context
}
