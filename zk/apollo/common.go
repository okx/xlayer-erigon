package apollo

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/urfave/cli/v2"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/erigon/turbo/node"
	"github.com/ledgerwatch/log/v3"
)

func (c *Client) unmarshal(value interface{}) (*nodecfg.Config, *ethconfig.Config, error) {
	mockCtx := cli.NewContext(nil, nil, nil)
	err := setFlagsFromBytes(mockCtx, value)
	if err != nil {
		log.Error(fmt.Sprintf("failed to set flags from bytes: %v", err))
		return nil, nil, err
	}

	nodeCfg := node.NewNodConfigUrfave(mockCtx)
	ethCfg := node.NewEthConfigUrfave(mockCtx, nodeCfg)

	return nodeCfg, ethCfg, nil
}

const (
	// Halt is the key for L2GasPricer halt
	Halt         = "Halt"
	maxHaltDelay = 20
)

func (c *Client) fireHalt(key string, value *storage.ConfigChange) {
	switch key {
	case Halt:
		if value.OldValue.(string) != value.NewValue.(string) {
			random, _ := rand.Int(rand.Reader, big.NewInt(maxHaltDelay))
			delay := time.Second * time.Duration(random.Int64())
			log.Info(fmt.Sprintf("halt changed from %s to %s delay halt %v", value.OldValue.(string), value.NewValue.(string), delay))
			time.Sleep(delay)
			os.Exit(1)
		}
	}
}
