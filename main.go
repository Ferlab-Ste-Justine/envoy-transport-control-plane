package main

import (
	"context"
	"fmt"
	"net"

	"ferlab/envoy-transport-control-plane/callbacks"
	"ferlab/envoy-transport-control-plane/config"
	"ferlab/envoy-transport-control-plane/logger"
	"ferlab/envoy-transport-control-plane/parameters"
	"ferlab/envoy-transport-control-plane/snapshot"
	"ferlab/envoy-transport-control-plane/utils"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
)

func GetGrpcServer(conf config.Config) *grpc.Server {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(1000000),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    conf.Server.KeepAliveTime,
			Timeout: conf.Server.KeepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             conf.Server.KeepAliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	return grpc.NewServer(grpcOptions...)
}

func main() {
	log := logger.Logger{LogLevel: logger.ERROR}
	
	conf, confErr := config.GetConfig("config.yml")
	utils.AbortOnErr(confErr)

	log.LogLevel = conf.GetLogLevel()

	params := parameters.GetParameters()

	ca := cache.NewSnapshotCache(true, cache.IDHash{}, log)
	snap, ssErr := snapshot.GetSnapshot(params)
	utils.AbortOnErr(ssErr)

	consErr := snap.Consistent()
	utils.AbortOnErr(consErr)

	setErr := ca.SetSnapshot(context.Background(), "test-id", snap)
	utils.AbortOnErr(setErr)

	srv := server.NewServer(context.Background(), ca, &callbacks.Callbacks{Logger: log})
	
	gsrv := GetGrpcServer(conf)
	lis, lisErr := net.Listen("tcp", fmt.Sprintf("%s:%d", conf.Server.BindIp, conf.Server.Port))
	utils.AbortOnErr(lisErr)

	clusterservice.RegisterClusterDiscoveryServiceServer(gsrv, srv)
	listenerservice.RegisterListenerDiscoveryServiceServer(gsrv, srv)

	log.Infof("Control plane server listening on %s:%d\n", conf.Server.BindIp, conf.Server.Port)
	srvErr := gsrv.Serve(lis)
	utils.AbortOnErr(srvErr)
}