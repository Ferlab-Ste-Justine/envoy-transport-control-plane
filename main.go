package main

import (
	"os"
	"os/signal"
	"syscall"

	"ferlab/envoy-transport-control-plane/config"
	"ferlab/envoy-transport-control-plane/logger"
	"ferlab/envoy-transport-control-plane/parameters"
	"ferlab/envoy-transport-control-plane/server"
	"ferlab/envoy-transport-control-plane/utils"
)

func getConfigFilePath() string {
	path := os.Getenv("ENVOY_TCP_CONFIG_FILE")
	if path == "" {
		return "config.yml"
	}
	return path
}

func main() {
	log := logger.Logger{LogLevel: logger.ERROR}

	conf, confErr := config.GetConfig(getConfigFilePath())
	utils.AbortOnErr(confErr, log)

	log.LogLevel = conf.GetLogLevel()

	paramsRetriever := parameters.Retriever{Logger: log, VersionFallback: conf.VersionFallback}
	paramsChan, paramsCancel := paramsRetriever.RetrieveParameters(conf, log)

	ca, caErrChan := server.SetCache(
		paramsChan,
		log,
	)
	serveCancel, serveErrChan := server.Serve(ca, conf, log)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Warnf("[main] Caught signal %s. Terminating.", sig.String())
		paramsCancel()
		serveCancel()
	}()

	select {
	case caErr := <-caErrChan:
		paramsCancel()
		serveCancel()
		<-serveErrChan
		<-paramsChan
		utils.AbortOnErr(caErr, log)
	case serverErr := <-serveErrChan:
		paramsCancel()
		serveCancel()
		<-paramsChan
		utils.AbortOnErr(serverErr, log)
	}
}
