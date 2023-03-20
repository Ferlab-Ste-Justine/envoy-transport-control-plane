package callbacks

import (
	"context"

	"ferlab/envoy-transport-control-plane/logger"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
)

type Callbacks struct {
	Logger         logger.Logger
}

func (cb *Callbacks) OnStreamOpen(_ context.Context, streamID int64, typeURL string) error {
	cb.Logger.Infof("Stream %d opened for resource %s\n", streamID, typeURL)
	return nil
}

func (cb *Callbacks) OnStreamClosed(streamID int64, node *core.Node) {
	cb.Logger.Infof("Stream %d closed for node %s\n", streamID, node.Id)
}

func (cb *Callbacks) OnDeltaStreamOpen(_ context.Context, streamID int64, typeURL string) error {
	cb.Logger.Infof("Delta stream %d opened for resource %s\n", streamID, typeURL)
	return nil
}

func (cb *Callbacks) OnDeltaStreamClosed(streamID int64, node *core.Node) {
	cb.Logger.Infof("Delta stream %d closed for node %s\n", streamID, node.Id)
}

func (cb *Callbacks) OnStreamRequest(streamID int64, req *discovery.DiscoveryRequest) error {
	cb.Logger.Infof("For stream %d, node %s requested resource %s. Version of last response was %s\n", streamID, req.Node.Id, req.TypeUrl, req.VersionInfo)
	return nil
}

func (cb *Callbacks) OnStreamResponse(_ context.Context, streamID int64, req *discovery.DiscoveryRequest, res *discovery.DiscoveryResponse) {
	cb.Logger.Infof("For stream %d on node %s with resource %s, server responded with version %s\n", streamID, req.Node.Id, req.TypeUrl, res.VersionInfo)
}

func (cb *Callbacks) OnStreamDeltaRequest(streamID int64, req *discovery.DeltaDiscoveryRequest) error {
	cb.Logger.Infof("For delta stream %d, node %s requested resource %s\n", streamID, req.Node.Id, req.TypeUrl)
	return nil
}

func (cb *Callbacks) OnStreamDeltaResponse(streamID int64, req *discovery.DeltaDiscoveryRequest, res *discovery.DeltaDiscoveryResponse) {
	cb.Logger.Infof("For delta stream %d on node %s with resource %s, server responded with version %s\n", streamID, req.Node.Id, req.TypeUrl, res.SystemVersionInfo)
}

func (cb *Callbacks) OnFetchRequest(_ context.Context, req *discovery.DiscoveryRequest) error {
	cb.Logger.Infof("Node %s requested resource %s. Version of last response was %s\n", req.Node.Id, req.TypeUrl)
	return nil
}

func (cb *Callbacks) OnFetchResponse(req *discovery.DiscoveryRequest, res *discovery.DiscoveryResponse) {
	cb.Logger.Infof("For node %s with resource %s, server responded with version %s\n", req.Node.Id, req.TypeUrl, res.VersionInfo)
}