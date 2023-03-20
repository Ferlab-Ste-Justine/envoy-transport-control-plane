package parameters

import (
	"time"

	/*yaml "gopkg.in/yaml.v2"*/
)

type DnsServer struct {
	Ip   string
	Port uint32
}

type HealthCheck struct {
	Timeout            time.Duration
	Interval           time.Duration
	HealthyThreshold   uint32        `yaml:"healthy_threshold"`
	UnhealthyThreshold uint32        `yaml:"unhealthy_threshold"`
}

type ExposedService struct {
	Name           string
	ListeningPort  uint32        `yaml:"listening_port"`
	ListeningIp    string        `yaml:"listening_ip"`
	ClusterDomain  string        `yaml:"cluster_domain"`
	ClusterPort    uint32        `yaml:"cluster_port"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxConnections uint64        `yaml:"max_connections"`
	HealthCheck    HealthCheck   `yaml:"health_check"`
}

type Parameters struct {
	DnsServers []DnsServer      `yaml:"dns_servers"`
    Services   []ExposedService
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
		Services: []ExposedService{
			ExposedService{
				Name: "server1",
				ListeningPort: 9081,
				ListeningIp: "127.0.0.1",
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
				ListeningIp: "127.0.0.1",
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
				ListeningIp: "127.0.0.1",
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