package ethconfig

import (
	"fmt"
)

const (
	DefaultMode = "default"
	RPCMode     = "rpc"
	SeqMode     = "sequencer"
)

func ValidateOptimizationMode(mode string) error {
	switch mode {
	case DefaultMode, RPCMode, SeqMode:
		return nil
	default:
		return fmt.Errorf("invalid optimization mode: %s. Must be one of: %s, %s, %s",
			mode, DefaultMode, RPCMode, SeqMode)
	}
}

func ApplyOptimizationModeConfig(cfg *Zk, optimizationMode string) {
	ValidateOptimizationMode(optimizationMode)
	switch optimizationMode {
	case RPCMode:
		cfg.Optimizations = RPCModeConfig
	case SeqMode:
		cfg.Optimizations = SequencerModeConfig
	default:
		cfg.Optimizations = DefaultOptimizationConfig
	}

}
