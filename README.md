# About

This is a control plane implementation for simple envoy load balancers that have one-to-one mappings between listeners and clusters with connection health checks. It assumes that discovery on the upstream servers will happen via dns resolution.

It fetches the envoy configuration from an etcd cluster with the expected workflow being that the keys in the etcd cluster will be updated with the desired envoy configuration which will propagate through this control plane to the envoy load balancers.

The control plane supports the specification of multiple load balancer configurations in the etcd keystore and could theoretically serve non-identical load balancers concurrently. However, tls is currently not implemented (as our immediate use-case is to update a single envoy server via localhost traffic) and serving several load balancers over an insecure network would be succeptible to man in the middle attacks.

# Usage

## Configuration

...

## Etcd Keyspace Format

...
