package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"ferlab/k8-lb-cp/callbacks"
	"ferlab/k8-lb-cp/logger"
	"ferlab/k8-lb-cp/parameters"
	"ferlab/k8-lb-cp/snapshot"
	"ferlab/k8-lb-cp/utils"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
)

func GetGrpcServer() *grpc.Server {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(1000000),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 5 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             30 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	return grpc.NewServer(grpcOptions...)
}

func main() {
	log := logger.Logger{LogLevel: logger.INFO}
	params := parameters.GetParameters()

	ca := cache.NewSnapshotCache(true, cache.IDHash{}, log)
	snap, ssErr := snapshot.GetSnapshot(params)
	utils.AbortOnErr(ssErr)

	consErr := snap.Consistent()
	utils.AbortOnErr(consErr)

	setErr := ca.SetSnapshot(context.Background(), "test-id", snap)
	utils.AbortOnErr(setErr)

	srv := server.NewServer(context.Background(), ca, &callbacks.Callbacks{Logger: log})
	
	gsrv := GetGrpcServer()
	lis, lisErr := net.Listen("tcp", fmt.Sprintf(":%d", 18000))
	utils.AbortOnErr(lisErr)

	clusterservice.RegisterClusterDiscoveryServiceServer(gsrv, srv)
	listenerservice.RegisterListenerDiscoveryServiceServer(gsrv, srv)

	log.Infof("Control plane server listening on %d\n", 18000)
	srvErr := gsrv.Serve(lis)
	utils.AbortOnErr(srvErr)

	//log.Infof("Test")
}