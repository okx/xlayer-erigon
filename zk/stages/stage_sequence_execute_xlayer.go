package stages

import (
	"fmt"
	"time"

	"github.com/ledgerwatch/erigon/zk/apollo"
	"github.com/ledgerwatch/log/v3"
)

func tryToSleepSequencer(localDuration time.Duration, logPrefix string) {
	fullBatchSleepDuration := apollo.GetFullBatchSleepDuration(localDuration)
	if fullBatchSleepDuration > 0 {
		log.Info(fmt.Sprintf("[%s] Slow down sequencer: %v", logPrefix, fullBatchSleepDuration))
		time.Sleep(fullBatchSleepDuration)
	}
}
