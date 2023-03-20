package parameters

import "time"

type DnsServer struct {
	Ip   string
	Port uint32
}

type HealthCheck struct {
	Timeout            time.Duration
	Interval           time.Duration
	HealthyThreshold   uint32
	UnhealthyThreshold uint32
}

type ExposedService struct {
	Name           string
	ListeningPort  uint32
	ClusterDomain  string
	ClusterPort    uint32
	IdleTimeout    time.Duration
	MaxConnections uint64
	HealthCheck   HealthCheck
}

type Parameters struct {
	DnsServers    []DnsServer
	BindingIp     string
    Services      []ExposedService
}

func GetParameters() Parameters {
	timeout, _ := time.ParseDuration("600s")
	hTimeout, _ := time.ParseDuration("10s")
	hInterval, _ := time.ParseDuration("30s")
	return Parameters{
		DnsServers: []DnsServer{
			DnsServer{
				Ip: "127.0.0.1",
				Port: 1053,
			},
		},
		BindingIp: "127.0.0.1",
		Services: []ExposedService{
			ExposedService{
				Name: "server1",
				ListeningPort: 9081,
				ClusterDomain: "test.local",
				ClusterPort: 8081,
				IdleTimeout: timeout,
				MaxConnections: 100,
				HealthCheck: HealthCheck{
					Timeout: hTimeout,
					Interval: hInterval,
					HealthyThreshold: 1,
					UnhealthyThreshold: 3,
				},
			},
			ExposedService{
				Name: "server2",
				ListeningPort: 9082,
				ClusterDomain: "test.local",
				ClusterPort: 8082,
				IdleTimeout: timeout,
				MaxConnections: 100,
				HealthCheck: HealthCheck{
					Timeout: hTimeout,
					Interval: hInterval,
					HealthyThreshold: 1,
					UnhealthyThreshold: 3,
				},
			},
			ExposedService{
				Name: "server3",
				ListeningPort: 9083,
				ClusterDomain: "test.local",
				ClusterPort: 8083,
				IdleTimeout: timeout,
				MaxConnections: 100,
				HealthCheck: HealthCheck{
					Timeout: hTimeout,
					Interval: hInterval,
					HealthyThreshold: 1,
					UnhealthyThreshold: 3,
				},
			},
		},
	}
}