package main

import (
	"fmt"
	"os"

	_ "github.com/ledgerwatch/erigon/core/snaptype"        //hack
	_ "github.com/ledgerwatch/erigon/polygon/bor/snaptype" //hack

	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon/cmd/integration/commands"
)

func main() {
	rootCmd := commands.RootCommand()
	fmt.Println("zjg----1")
	ctx, _ := common.RootContext()
	fmt.Println("zjg----2")
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Println("zjg----3")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("zjg----4")
}
