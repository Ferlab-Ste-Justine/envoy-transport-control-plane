package main

import (
	"ferlab/envoy-transport-control-plane/config"
	"ferlab/envoy-transport-control-plane/logger"
	"ferlab/envoy-transport-control-plane/parameters"
	"ferlab/envoy-transport-control-plane/server"
	"ferlab/envoy-transport-control-plane/utils"
)

func main() {
	log := logger.Logger{LogLevel: logger.ERROR}
	
	conf, confErr := config.GetConfig("config.yml")
	utils.AbortOnErr(confErr, log)

	log.LogLevel = conf.GetLogLevel()

	paramsRetriever := parameters.Retriever{Logger: log, VersionFallback: conf.VersionFallback}
	paramsChan, paramsCancel := paramsRetriever.RetrieveParameters(conf, log)

	ca, caErrChan := server.SetCache(
		paramsChan, 
		log,
	)
	serveCancel, serveErrChan := server.Serve(ca, conf, log)

	select {
	case caErr := <- caErrChan:
		paramsCancel()
		serveCancel()
		utils.AbortOnErr(caErr, log)
	case serverErr := <- serveErrChan:
		paramsCancel()
		serveCancel()
		utils.AbortOnErr(serverErr, log)
	}
}