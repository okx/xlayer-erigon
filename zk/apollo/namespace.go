package apollo

const (
	// JsonRPC is the jsonrpc prefix namespace, the content of the prefix is the configuration for jsonrpc with yaml format
	JsonRPC = "jsonrpc"
	// Sequencer is the sequencer prefix namespace, the content of the prefix is the configuration for sequencer with yaml format
	Sequencer = "sequencer"
	// L2GasPricer is the l2gaspricer prefix namespace, the content of the prefix is the configuration for l2gaspricer with yaml format
	L2GasPricer = "l2gaspricer"
	// Halt is the halt suffix namespace. Change the halt to a different value will halt the respective service
	Halt = "halt"
)
