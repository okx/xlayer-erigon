package main

import (
	"fmt"
	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/zk/apollo"
	"github.com/ledgerwatch/erigon/zk/metrics"
	"github.com/ledgerwatch/log/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"time"
)

const (
	MetricsEndpoint = "/metrics"
)

func initRunForXLayer(ethCfg *ethconfig.Config) {
	apolloClient := apollo.NewClient(ethCfg)
	if apolloClient.LoadConfig() {
		log.Info("Apollo config loaded")
	}

	// Start Metrics Server
	if ethCfg.Zk.XLayer.Metrics.Enabled {
		metrics.XLayerMetricsInit()
		go startMetricsHttpServer(ethCfg.Zk.XLayer.Metrics)
	}
}

func startMetricsHttpServer(c ethconfig.MetricsConfig) {
	const ten = 10
	mux := http.NewServeMux()
	address := fmt.Sprintf("%s:%d", c.Host, c.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("failed to create tcp listener for metrics", "err", err)
		return
	}
	mux.Handle(MetricsEndpoint, promhttp.Handler())

	metricsServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: ten * time.Second,
		ReadTimeout:       ten * time.Second,
	}
	log.Info("metrics server listening on port", c.Port)
	if err := metricsServer.Serve(lis); err != nil {
		if err == http.ErrServerClosed {
			log.Warn("http server for metrics stopped")
			return
		}
		log.Error("closed http connection for metrics server", "err", err)
		return
	}
}
