package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ledgerwatch/erigon/eth/ethconfig"
	"github.com/ledgerwatch/erigon/zk/apollo"
	"github.com/ledgerwatch/erigon/zk/metrics"
	"github.com/ledgerwatch/log/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	MetricsEndpoint             = "/metrics"
	MetricsServerDefaultTimeOut = 10
)

func initRunForXLayer(ethCfg *ethconfig.Config) {
	apolloClient := apollo.NewClient(ethCfg)
	if apolloClient.LoadConfig() {
		log.Info("Apollo config loaded")
	}

	// Start Metrics Server
	if ethCfg.Zk.XLayer.Metrics.Enabled {
		go startMetricsHttpServer(ethCfg.Zk.XLayer.Metrics)
	}
}

func startMetricsHttpServer(c ethconfig.MetricsConfig) {
	metrics.XLayerMetricsInit()
	mux := http.NewServeMux()
	address := fmt.Sprintf("%s:%d", c.Host, c.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Error(fmt.Sprintf("failed to create %v for metrics, error:%v", address, err))
		return
	}
	mux.Handle(MetricsEndpoint, promhttp.Handler())

	metricsServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: MetricsServerDefaultTimeOut * time.Second,
		ReadTimeout:       MetricsServerDefaultTimeOut * time.Second,
	}
	log.Info("metrics server listening on port", c.Port)
	if err := metricsServer.Serve(lis); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Warn("http server for metrics stopped")
			return
		}
		log.Error(fmt.Sprintf("closed http connection %v for metrics server, error:%v", address, err))
		return
	}
}
