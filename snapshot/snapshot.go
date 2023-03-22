package snapshot

import (
	"ferlab/envoy-transport-control-plane/parameters"

	"fmt"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	stream "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/stream/v3"
	tcpproxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	connlimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/connection_limit/v3"
	cares "github.com/envoyproxy/go-control-plane/envoy/extensions/network/dns_resolver/cares/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func getCluster(service parameters.ExposedService, dnsServers []parameters.DnsServer) (*cluster.Cluster, error) {
	dnsResolvers := []*core.Address{}
	for _, dnsServer := range dnsServers {
		dnsResolvers = append(dnsResolvers, &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Address: dnsServer.Ip,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: dnsServer.Port,
					},
				},
			},
		})
	}
	dnsResolverConfig, err := anypb.New(&cares.CaresDnsResolverConfig{
		Resolvers: dnsResolvers,
		UseResolversAsFallback: false,
	})
	if err != nil {
		return nil, err
	}
	
	return &cluster.Cluster{
		Name: service.Name,
		ClusterDiscoveryType: &cluster.Cluster_Type{
			Type: cluster.Cluster_STRICT_DNS,
		},
		TypedDnsResolverConfig: &core.TypedExtensionConfig{
			Name: "envoy.typed_dns_resolver_config",
			TypedConfig: dnsResolverConfig,
		},
		LbPolicy: cluster.Cluster_ROUND_ROBIN,
		HealthChecks: []*core.HealthCheck{
			&core.HealthCheck{
				Timeout: &durationpb.Duration{
					Seconds: service.HealthCheck.Timeout.Nanoseconds() / 1000000000,
					Nanos: int32(service.HealthCheck.Timeout.Nanoseconds() - service.HealthCheck.Timeout.Round(time.Second).Nanoseconds()),
				},
				Interval: &durationpb.Duration{
					Seconds: service.HealthCheck.Interval.Nanoseconds() / 1000000000,
					Nanos: int32(service.HealthCheck.Interval.Nanoseconds() - service.HealthCheck.Interval.Round(time.Second).Nanoseconds()),
				},
				HealthyThreshold: &wrapperspb.UInt32Value{Value: service.HealthCheck.HealthyThreshold},
				UnhealthyThreshold: &wrapperspb.UInt32Value{Value: service.HealthCheck.UnhealthyThreshold},
				ReuseConnection: &wrapperspb.BoolValue{Value: false},
				HealthChecker: &core.HealthCheck_TcpHealthCheck_{
					TcpHealthCheck: &core.HealthCheck_TcpHealthCheck{},
				},
			},
		},
		CircuitBreakers: &cluster.CircuitBreakers{
			Thresholds: []*cluster.CircuitBreakers_Thresholds{
				&cluster.CircuitBreakers_Thresholds{
					Priority: core.RoutingPriority_DEFAULT,
					MaxConnections: &wrapperspb.UInt32Value{Value: uint32(service.MaxConnections)},
					MaxPendingRequests: &wrapperspb.UInt32Value{Value: uint32(service.MaxConnections)},
					MaxRequests: &wrapperspb.UInt32Value{Value: uint32(service.MaxConnections)},
					MaxRetries: &wrapperspb.UInt32Value{Value: 3},
				},
			},
		},
		LoadAssignment: &endpoint.ClusterLoadAssignment{
			ClusterName: service.Name,
			Endpoints: []*endpoint.LocalityLbEndpoints{
				&endpoint.LocalityLbEndpoints{
					LbEndpoints: []*endpoint.LbEndpoint{
						&endpoint.LbEndpoint{
							HostIdentifier: &endpoint.LbEndpoint_Endpoint{
								Endpoint: &endpoint.Endpoint{
									Address: &core.Address{
										Address: &core.Address_SocketAddress{
											SocketAddress: &core.SocketAddress{
												Protocol: core.SocketAddress_TCP,
												Address:  service.ClusterDomain,
												PortSpecifier: &core.SocketAddress_PortValue{
													PortValue: service.ClusterPort,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func getListener(service parameters.ExposedService, dnsServers []parameters.DnsServer) (*listener.Listener, error) {
	connLimit, err := anypb.New(&connlimit.ConnectionLimit{
		StatPrefix: fmt.Sprintf("%s_listener_connection_limit", service.Name),
		MaxConnections: &wrapperspb.UInt64Value{Value: service.MaxConnections},
	})
	if err != nil {
		return nil, err
	}

	stdoutLogs, lErr := anypb.New(&stream.StdoutAccessLog{
		AccessLogFormat: &stream.StdoutAccessLog_LogFormat{
			LogFormat: &core.SubstitutionFormatString{
				Format: &core.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &core.DataSource{
						Specifier: &core.DataSource_InlineString{
							InlineString: service.AccessLogFormat,
						},
					},
				},
			},
		},
	})
	if lErr != nil {
		return nil, lErr
	}

	tcpProxy, err := anypb.New(&tcpproxy.TcpProxy{
		StatPrefix: fmt.Sprintf("%s_listener_tcp_proxy", service.Name),
		ClusterSpecifier: &tcpproxy.TcpProxy_Cluster{service.Name},
		IdleTimeout: &durationpb.Duration{
			Seconds: service.IdleTimeout.Nanoseconds() / 1000000000,
			Nanos: int32(service.IdleTimeout.Nanoseconds() - service.IdleTimeout.Round(time.Second).Nanoseconds()),
		},
		AccessLog: []*accesslog.AccessLog{
			&accesslog.AccessLog{
				Name: fmt.Sprintf("%s_listener_tcp_log", service.Name),
				ConfigType: &accesslog.AccessLog_TypedConfig{
					TypedConfig: stdoutLogs,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	
	return &listener.Listener{
		Name: service.Name,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address: service.ListeningIp,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: service.ListeningPort,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{
				{
					Name: "envoy.filters.network.connection_limit",
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: connLimit,
					},
				},
				{
					Name: "envoy.filters.network.tcp_proxy",
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: tcpProxy,
					},
				},
		    },
		}},
	}, nil
}

func GetSnapshot(params parameters.Parameters) (*cache.Snapshot, error) {
	resources := map[resource.Type][]types.Resource{
		resource.ClusterType: []types.Resource{},
		resource.ListenerType: []types.Resource{},
	}

	for _, service := range params.Services {
		clust, clustErr := getCluster(service, params.DnsServers)
		if clustErr != nil {
			return nil, clustErr
		}

		list, listErr := getListener(service, params.DnsServers)
		if listErr != nil {
			return nil, listErr
		}

		resources[resource.ClusterType] = append(
			resources[resource.ClusterType],
			clust,
		)
		resources[resource.ListenerType] = append(
			resources[resource.ListenerType],
			list,
		)
	}

	snap, snErr := cache.NewSnapshot(
		params.Version, 
		resources,
	)
	return snap, snErr
}