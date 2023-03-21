package parameters

import (
	"strings"
	"time"

	"ferlab/envoy-transport-control-plane/config"
	"ferlab/envoy-transport-control-plane/logger"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
	yaml "gopkg.in/yaml.v2"
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

type NodeParameters struct {
	NodeId     string
	Delete     bool
	Parameters Parameters
}

type NodeParametersRetrieval struct {
	NodeParameters NodeParameters
	Error          error
}

type Retriever struct {
	Client *client.EtcdClient
}

func (r *Retriever) getPrefixNodeParams(prefix string) ([]NodeParameters, int64, error) {
	result := []NodeParameters{}
	
	info, revision, err := r.Client.GetPrefix(prefix)
	if err != nil {
		return result, revision, err
	}

	for _, val := range info {
		var params Parameters
		key := strings.TrimPrefix(val.Key, prefix)
		nodeId := strings.Split(key, "/")[0]
		err = yaml.Unmarshal([]byte(val.Value), &params)
		if err != nil {
			return result, revision, err
		}
		result = append(result, NodeParameters{
			NodeId: nodeId,
			Delete: false,
			Parameters: params,
		})
	}

	return result, revision, nil
}

func (r *Retriever) watchPrefixNodeParams(prefix string, revision int64, retrievalChan chan<- NodeParametersRetrieval) {
	changesChan := r.Client.WatchPrefixChanges(prefix, revision)

	for change := range changesChan {
		if change.Error != nil {
			retrievalChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{}, Error: change.Error}
			return
		}

		for _, key := range change.Changes.Deletions {
			key = strings.TrimPrefix(key, prefix)
			nodeId := strings.Split(key, "/")[0]
			retrievalChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{
				NodeId: nodeId,
				Delete: true,
			}, Error: nil}
		}

		for key, val := range change.Changes.Upserts {
			var params Parameters
			key = strings.TrimPrefix(key, prefix)
			nodeId := strings.Split(key, "/")[0]
			err := yaml.Unmarshal([]byte(val), &params)
			if err != nil {
				retrievalChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{}, Error: err}
				return
			}
			retrievalChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{
				NodeId: nodeId,
				Delete: false,
				Parameters: params,
			}, Error: nil}
		}
	}
}

func (r *Retriever) RetrieveParameters(conf config.Config, log logger.Logger) (chan NodeParametersRetrieval) {
	paramsChan := make(chan NodeParametersRetrieval)

	go func() {
		defer close(paramsChan)
		cli, cliErr := client.Connect(client.EtcdClientOptions{
			ClientCertPath:    conf.EtcdClient.Auth.ClientCert,
			ClientKeyPath:     conf.EtcdClient.Auth.ClientKey,
			CaCertPath:        conf.EtcdClient.Auth.CaCert,
			Username:          conf.EtcdClient.Auth.Username,
			Password:          conf.EtcdClient.Auth.Password,
			EtcdEndpoints:     conf.EtcdClient.Endpoints,
			ConnectionTimeout: conf.EtcdClient.ConnectionTimeout,
			RequestTimeout:    conf.EtcdClient.RequestTimeout,
			Retries:           conf.EtcdClient.Retries,
		})
		if cliErr != nil {
			paramsChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{}, Error: cliErr}
			return
		}

		r.Client = cli

		nodesParams, revision, nodeParamsErr := r.getPrefixNodeParams(conf.EtcdClient.Prefix)
		if nodeParamsErr != nil {
			paramsChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{}, Error: nodeParamsErr}
			return
		}

		for _, nodeParams := range nodesParams {
			paramsChan <- NodeParametersRetrieval{NodeParameters: nodeParams, Error: nil}
		}

		r.watchPrefixNodeParams(conf.EtcdClient.Prefix, revision, paramsChan)
	}()

	return paramsChan
}