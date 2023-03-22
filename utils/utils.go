package utils

import (
	"os"

	"ferlab/envoy-transport-control-plane/logger"
)

func AbortOnErr(err error, log logger.Logger) {
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
}
