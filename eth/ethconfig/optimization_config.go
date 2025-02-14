package ethconfig

// OptimizationConfig defines optimization configurations for different modes
type OptimizationConfig struct {
	// NewOptimization bool
}

var (
	// RPCModeConfig defines optimizations for RPC mode
	RPCModeConfig = OptimizationConfig{
		// NewOptimization: false,
	}

	// SequencerModeConfig defines optimizations for sequencer mode
	SequencerModeConfig = OptimizationConfig{
		// NewOptimization: true,
	}

	DefaultOptimizationConfig = OptimizationConfig{
		// NewOptimization: false,
	}
)
