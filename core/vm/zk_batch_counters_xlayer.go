package vm

import (
	"fmt"
	"github.com/ledgerwatch/log/v3"
)

// CheckOkPayForOverflow returns true in the case that any counter has less than 0 remaining
func (bcc *BatchCounterCollector) CheckOkPayForOverflow(okPayCounterLimitPercentage uint) (bool, error) {
	combined := bcc.NewCounters()
	for k, _ := range combined {
		val := bcc.rlpCombinedCounters[k].used + bcc.executionCombinedCounters[k].used + bcc.processingCombinedCounters[k].used
		combined[k].used += val
		combined[k].remaining -= val
	}

	overflow := false
	for _, v := range combined {
		if v.initialAmount*int(okPayCounterLimitPercentage)/100 < v.used {
			log.Info("[VCOUNTER] OkPay Counter overflow detected", "counter", v.name, "remaining", v.remaining, "used", v.used)
			overflow = true
		}
	}

	// if we have an overflow we want to log the counters for debugging purposes
	if overflow {
		logText := "[VCOUNTER] Counters stats"
		for _, v := range combined {
			logText += fmt.Sprintf(" %s: initial: %v used: %v (remaining: %v)", v.name, v.initialAmount, v.used, v.remaining)
		}
		log.Info(logText)
	}

	return overflow, nil
}
