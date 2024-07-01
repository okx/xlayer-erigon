package apollo

import (
	"flag"
	"fmt"
	"strings"

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

func getNamespacePrefix(namespace string) (string, error) {
	items := strings.Split(namespace, "-")
	if len(items) < NamespaceSplits {
		return "", fmt.Errorf("invalid namespace: %s, no separator \"-\" present, please configure apollo namespace in the correct format \"prefix-item\"", namespace)
	}
	return items[0], nil
}

func getNamespaceSuffix(namespace string) (string, error) {
	items := strings.Split(namespace, "-")
	if len(items) < NamespaceSplits {
		return "", fmt.Errorf("invalid namespace: %s, no separator \"-\" present, please configure apollo namespace in the correct format \"item-suffix\"", namespace)
	}
	return items[len(items)-1], nil
}
