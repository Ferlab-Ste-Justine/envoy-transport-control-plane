package parameters

import (
	"context"
	"errors"
	"fmt"
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
	Http               HttpHealthCheck
}

type StatusCodeRange struct {
	First int64
	Last int64
}

type HttpHealthCheck struct {
	Enabled         bool
	Path            string
	StatusCodeRange StatusCodeRange  `yaml:"status_code_range"`
}

type HttpListener struct {
	Enabled                     bool
	ServerName                  string          `yaml:"server_name"`         
	MaxConcurrentStreams        uint32          `yaml:"max_concurrent_streams"`
	RequestHeadersTimeout       time.Duration   `yaml:"request_headers_timeout"`
	UseRemoteAddress            bool            `yaml:"use_remote_address"`
	InitialConnectionWindowSize uint32          `yaml:"initial_connection_window_size"`
	InitialStreamWindowSize     uint32          `yaml:"initial_stream_window_size"`
}

type TlsTermination struct {
	ListenerCertificate      string       `yaml:"listener_certificate"`
	ListenerKey              string       `yaml:"listener_key"`
	ClusterCaCertificate     string       `yaml:"cluster_ca_certificate"`
	ClusterClientCertificate string       `yaml:"cluster_client_certificate"`
	ClusterClientKey         string       `yaml:"cluster_client_key"`
	HttpListener             HttpListener `yaml:"http_listener"`
}

type ExposedService struct {
	Name                 string
	ListeningPort        uint32         `yaml:"listening_port"`
	ListeningIp          string         `yaml:"listening_ip"`
	ClusterDomain        string         `yaml:"cluster_domain"`
	ClusterPort          uint32         `yaml:"cluster_port"`
	IdleTimeout          time.Duration  `yaml:"idle_timeout"`
	MaxConnections       uint64         `yaml:"max_connections"`
	HealthCheck          HealthCheck    `yaml:"health_check"`
	AccessLogFormat      string         `yaml:"access_log_format"`
	TlsTermination       TlsTermination `yaml:"tls_termination"`
}

type Parameters struct {
	Version    string
	DnsServers []DnsServer `yaml:"dns_servers"`
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
	Logger          logger.Logger
	VersionFallback string
	Client          *client.EtcdClient
}

func (r *Retriever) setParamsVersion(params *Parameters, etcdVer int64) error {
	if params.Version != "" {
		return nil
	}

	if r.VersionFallback == "none" {
		return errors.New("Read parameters without a version and there is no version fallback strategy")
	}

	if r.VersionFallback == "etcd" {
		params.Version = fmt.Sprintf("%d", etcdVer)
	}

	if r.VersionFallback == "time" {
		params.Version = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	return nil
}

func (r *Retriever) getPrefixNodeParams(prefix string) ([]NodeParameters, int64, error) {
	result := []NodeParameters{}

	info, err := r.Client.GetPrefix(prefix)
	if err != nil {
		return result, info.Revision, err
	}

	for _, val := range info.Keys {
		nodeId := strings.TrimPrefix(val.Key, prefix)

		var params Parameters
		err = yaml.Unmarshal([]byte(val.Value), &params)
		if err != nil {
			return result, info.Revision, err
		}

		err = r.setParamsVersion(&params, val.Version)
		if err != nil {
			return result, info.Revision, err
		}

		r.Logger.Infof("[Etcd] Adding snapshot for node %s on boot", nodeId)
		result = append(result, NodeParameters{
			NodeId:     nodeId,
			Delete:     false,
			Parameters: params,
		})
	}

	return result, info.Revision, nil
}

func (r *Retriever) watchPrefixNodeParams(prefix string, revision int64, retrievalChan chan<- NodeParametersRetrieval) {
	changesChan := r.Client.Watch(prefix, client.WatchOptions{
		Revision: revision,
		IsPrefix: true,
		TrimPrefix: true,
	})

	r.Logger.Infof("[Etcd] Started watching for parameters updates")
	for change := range changesChan {
		if change.Error != nil {
			retrievalChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{}, Error: change.Error}
			return
		}

		for _, key := range change.Changes.Deletions {
			nodeId := strings.TrimPrefix(key, prefix)

			r.Logger.Infof("[Etcd] Removing snapshot for node %s on watch", nodeId)
			retrievalChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{
				NodeId: nodeId,
				Delete: true,
			}, Error: nil}
		}

		for key, val := range change.Changes.Upserts {
			nodeId := strings.TrimPrefix(key, prefix)

			var params Parameters
			err := yaml.Unmarshal([]byte(val.Value), &params)
			if err != nil {
				retrievalChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{}, Error: err}
				return
			}

			err = r.setParamsVersion(&params, val.Version)
			if err != nil {
				retrievalChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{}, Error: err}
				return
			}

			r.Logger.Infof("[Etcd] Adding/updating snapshot for node %s on watch", nodeId)
			retrievalChan <- NodeParametersRetrieval{NodeParameters: NodeParameters{
				NodeId:     nodeId,
				Delete:     false,
				Parameters: params,
			}, Error: nil}
		}
	}
	r.Logger.Infof("[Etcd] Watching for parameters updates stopped")
}

func (r *Retriever) RetrieveParameters(conf config.Config, log logger.Logger) (chan NodeParametersRetrieval, context.CancelFunc) {
	paramsChan := make(chan NodeParametersRetrieval)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(paramsChan)
		cli, cliErr := client.Connect(ctx, client.EtcdClientOptions{
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

		r.watchPrefixNodeParams(conf.EtcdClient.Prefix, revision+1, paramsChan)
	}()

	return paramsChan, cancel
}
