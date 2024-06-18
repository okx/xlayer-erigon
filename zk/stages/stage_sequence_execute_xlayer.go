package stages

import (
	"fmt"
	"time"

	"github.com/ledgerwatch/log/v3"
)

func tryHaltSequencer(logPrefix string, cfg SequenceBlockCfg, thisBatch uint64) {
	if cfg.zk.SequencerHaltOnBatchNumber == thisBatch {
		for {
			log.Info(fmt.Sprintf("[%s] Halt sequencer on batch %d...", logPrefix, thisBatch))
			time.Sleep(5 * time.Second) //nolint:gomnd
		}
	}
}
