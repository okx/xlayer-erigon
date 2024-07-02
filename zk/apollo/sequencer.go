package apollo

import (
	"fmt"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/ledgerwatch/erigon/cmd/utils"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/node/nodecfg"
	"github.com/ledgerwatch/log/v3"
)

func (c *Client) loadSequencer(value interface{}) {
	_, _, err := c.unmarshal(value)
	if err != nil {
		utils.Fatalf("load sequencer from apollo config failed, unmarshal err: %v", err)
	}

	// TODO: Add specific sequencer configs to load from apollo config
	log.Info(fmt.Sprintf("loaded sequencer from apollo config: %+v", value.(string)))
}

// fireSequencer fires the sequencer config change
func (c *Client) fireSequencer(key string, value *storage.ConfigChange) {
	nodeCfg, ethCfg, err := c.unmarshal(value.NewValue)
	if err != nil {
		log.Error(fmt.Sprintf("fire sequencer from apollo config failed, unmarshal err: %v", err))
		return
	}

	log.Info(fmt.Sprintf("apollo eth backend old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo eth backend config changed: %+v", value.NewValue.(string)))

	log.Info(fmt.Sprintf("apollo node old config : %+v", value.OldValue.(string)))
	log.Info(fmt.Sprintf("apollo node config changed: %+v", value.NewValue.(string)))

	nodecfg.UpdateSequencerConfig(*nodeCfg)
	ethconfig.UpdateSequencerConfig(*ethCfg)
}
