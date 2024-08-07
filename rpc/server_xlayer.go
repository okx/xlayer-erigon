package rpc

import "github.com/ledgerwatch/erigon/node/nodecfg"

// SetBatchEnabled sets if batches requests are handled by this server
func (s *Server) SetBatchEnabled(flag bool) {
	s.batchEnabled = flag
}

func (s *Server) getBatchReqLimitXLayer() (bool, int) {
	// if apollo is enabled, get the config from apollo
	if nodecfg.IsApolloConfigEnable() {
		if apolloConf, err := nodecfg.GetApolloConfig(); err == nil {
			return apolloConf.Http.BatchEnabled, apolloConf.Http.BatchLimit
		}
	}
	return s.batchEnabled, s.batchLimit
}
