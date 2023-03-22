package server

import (
	"context"
	"fmt"
	"net"

	"ferlab/envoy-transport-control-plane/callbacks"
	"ferlab/envoy-transport-control-plane/config"
	"ferlab/envoy-transport-control-plane/logger"
	"ferlab/envoy-transport-control-plane/parameters"
	"ferlab/envoy-transport-control-plane/snapshot"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func SetCache(paramsChan <-chan parameters.NodeParametersRetrieval, log logger.Logger) (*cache.SnapshotCache, <-chan error) {
	ca := cache.NewSnapshotCache(true, cache.IDHash{}, log)
	errChan := make(chan error)

	go func() {
		defer close(errChan)

		for params := range paramsChan {
			if params.Error != nil {
				errChan <- params.Error
				return
			}
			if !params.NodeParameters.Delete {
				snap, ssErr := snapshot.GetSnapshot(params.NodeParameters.Parameters)
				if ssErr != nil {
					errChan <- ssErr
					return
				}

				consErr := snap.Consistent()
				if consErr != nil {
					errChan <- consErr
					return
				}

				setErr := ca.SetSnapshot(context.Background(), params.NodeParameters.NodeId, snap)
				if setErr != nil {
					errChan <- setErr
					return
				}
			} else {
				ca.ClearSnapshot(params.NodeParameters.NodeId)
			}
		}

	}()

	return &ca, errChan
}

func GetGrpcServer(conf config.Config) *grpc.Server {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(conf.Server.MaxConnections),
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

type CancelServe func()

func Serve(ca *cache.SnapshotCache, conf config.Config, log logger.Logger) (CancelServe, chan error) {
	errChan := make(chan error)

	srv := server.NewServer(context.Background(), *ca, &callbacks.Callbacks{Logger: log})

	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(conf.Server.MaxConnections),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    conf.Server.KeepAliveTime,
			Timeout: conf.Server.KeepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             conf.Server.KeepAliveMinTime,
			PermitWithoutStream: true,
		}),
	)
	gsrv := grpc.NewServer(grpcOptions...)
	cancel := func() {
		gsrv.Stop()
	}

	go func() {
		defer close(errChan)

		lis, lisErr := net.Listen("tcp", fmt.Sprintf("%s:%d", conf.Server.BindIp, conf.Server.Port))
		if lisErr != nil {
			errChan <- lisErr
			return
		}

		clusterservice.RegisterClusterDiscoveryServiceServer(gsrv, srv)
		listenerservice.RegisterListenerDiscoveryServiceServer(gsrv, srv)

		log.Infof("[server] Listening on %s:%d", conf.Server.BindIp, conf.Server.Port)
		srvErr := gsrv.Serve(lis)
		if srvErr != nil {
			errChan <- srvErr
		}
		log.Infof("[server] Server stopped")
	}()

	return cancel, errChan
}
