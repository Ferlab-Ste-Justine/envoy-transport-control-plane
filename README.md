# About

This is a control plane implementation for simple envoy load balancers that have one-to-one mappings between listeners and clusters with connection health checks. It assumes that discovery on the upstream servers will happen via dns resolution.

It fetches the envoy configuration from an etcd cluster with the expected workflow being that the keys in the etcd cluster will be updated with the desired envoy configuration which will propagate through this control plane to the envoy load balancers.

The control plane supports the specification of multiple load balancer configurations in the etcd keystore and could theoretically serve non-identical load balancers concurrently. However, tls is currently not implemented (as our immediate use-case is to update a single envoy server via localhost traffic) and serving several load balancers over an insecure network would be succeptible to man in the middle attacks.

# Usage

## Configuration

The configuration file is in yaml and its path defaults to **config.yml** in the running directory. It can be configured to another path via the **ENVOY_TCP_CONFIG_FILE** environment variable.

The format of the configuration file is as follows:

```
etcd_client:
  prefix: Prefix in etcd that the control plane will search for configurations
  endpoints: List of etcd endpoints having the <address>:<port> format
  connection_timeout: Connection timeout to etcd in golang duration format
  request_timeout: Request timeout to etcd in golang duration format
  retries: Number of times to re-attempt failed connections/requests before aborting the process
  auth:
    ca_cert: Path to a CA certificate that will authentify the etcd servers certificates
    client_cert: Path to a client certificate that the control plane will use to authentify against etcd if client certificate authentication is used
    client_key: Path to a client key that the control plane will use to authentify against etcd if client certificate authentication is used
    password_auth: Path to a file in yaml format containing the username/password authentication credentials that the control plane will use to authentify to etcd if username/password authentication is used. The file should contain the following two keys: username, password
server:
  port: Port that the control plane will listen on for envoy server grpc requests
  bind_ip: Ip that the control plane will bind on for envoy server grpc requests
  max_connections: Maximum number of concurrent connections that the control plane will access
  keep_alive_time: Frequency at which the control plane should ping an envoy server to see if it is still alive if there is no traffic. Should be in golang duration format.
  keep_alive_timeout: Deadline after sending a keep alive ping after which the control plane will close the connection to an envoy server if there is no response. Should be in golang duration format. 
  keep_alive_min_time: Expected minimum amount of time before an envoy client sends a keep alive ping to the control plane. The control plane will close the connection if an envoy client violates this. 
log_level: Minimum log level to show. Can be: debug, info, warn, error
version_fallback: What to fallback on for snapshot versions if a version is not specified in the envoy config found in etcd. It can be: etcd (version of the key will be used), time (nanoseconds since epoch will be used) or none (error will be returned if the version is not defined)
```

## Etcd Keyspace Format

The keyspace for the etcd configuration is expected to be the following: `<search prefix><node id>` where the **node id** should match what is in the configuration of your envoy load balancers.

For example, if you have 2 envoy servers with node ids corresponding to **envoy-one** and **envoy-two** respectively and that your etcd search prefix is **/envoy-control-plane/**, then your configurations for envoy one and two should be found in the following etcd keys:

```
/envoy-control-plane/envoy-one
/envoy-control-plane/envoy-two
```

The configuration should be in yaml with the following format:

```
Version: Version of the configuration. Can be omitted in which case the control plane will use a fallback strategy to determine a version.
dns_servers:
  - ip: Ip of the first dns server
    port: Port of the first dns server
  ...
services:
  - name: Unique name of the first service
    listening_port: Port that envoy should listen on for the first service
    listening_ip: Ip that envoy should bind on for the first service
    cluster_domain: Domain that envoy should use to discover upstream servers   for the first service that it should forward requests to
    cluster_port: Port that envoy should forward requests to on the upstream servers for the first service
    idle_timeout: How long envoy should wait before closing a connection that has not traffic on the first service. Should be in goland duration format.
    max_connections: Maximum number of concurrent connections that envoy should accept for the first service
    healthCheck:
      timeout: Timeout on health checks on the first service endpoints
      interval: Interval of health checks on each of the first service endpoints
      healthy_threshold: Number of health checks that should pass on an unhealthy endpoing of the first service before it is deemed health
      unhealthy_threshold: Number of health checks that should failed on an healthy endpoint of the first service before it is deemed unhealthy
    access_log_format: Format for the access logs of the first service. See: https://www.envoyproxy.io/docs/envoy/v1.25.2/configuration/observability/access_log/usage#config-access-log-format-strings
    tls_termination: 
      listener_certificate: Path to a tls certificate to present to clients if the service should perform tls termination as opposed to tls passthrough. 
      listener_key: This field should be specified in combination with "listener_certificate". This should be the filesystem path to the private key that the service will use in combination with the certificate to authentify itself to the client.
      cluster_ca_certificate: Path to a CA certificate to authentify the backend certificate if the backend expects a tls connection as well.
      cluster_client_certificate: Path to a client certificate to present to the backend. If the backend doesn't validate client certs, a dummy certificate can be passed here.
      cluster_client_key: Path to a client key to present to the backend.
      use_http_listener: Whether to use an L7 http listener instead of L4 tcp as part of the tls termination. Useful to negotiate http version with clients.
      http_parameters: Parameters if using https listerner.
        server_name: Server name to return on http requests.
        max_concurrent_streams: Maximum number of streams to allow per peer for clients using http/2.
        request_headers_timeout: The maximum amount of time envoy will wait for all headers to be received once the initial byte is sent.
        use_remote_address: Whether envoy should append to the **x-forwarded-for** header with the client ip. Should be set to true when setting up an edge load balancer.
        initial_connection_window_size: Window size (in bytes) allocated to a new http/2 connection. Should be set to conservatively small amount for edge load balancer.
        initial_streaming_window_size: Window size (in bytes) allocated to a new http/2 stream. Should be set to conservatively small amount for edge load balancer.
  ...
```

## Test Environment

We setup a quick (but not minimalistic test environment) to troubleshoot quickly.

It requires the followig:
  - A recent version (1.25.2 at the time of this writing) of envoy's binary in your path
  - A recent version of coredns in your path
  - golang (version 1.18 or above)
  - terraform
  - A kubernetes running locally (example defaults to microk8s)

Simplifying the requirements of the test is not a high priority at this time, but we can get to it if there is a demand.

To run it:
- Start your dns server locally (run **test-setup/dns/run.sh**)
- Start etcd in kubernetes mapped with a nodeport (run **terraform init && terraform apply** in **test-setup/etcd**)
- Start your upstream server (run **test-setup/upstream-server/run.sh**)
- Start your envoy control plan (run **test-setup/control-plane/run.sh**)
- Load configurations for envoy one and two in etcd (run **terraform init && terraform apply** in **load balancer-configs**)
- Start your first envoy server (run **test-setup/envoy/run.sh**)
- Start your second envoy server (run **test-setup/envoy/run.sh**)

The 5 services of the upstream server will be listening directly on the following endpoints:
```
127.0.0.1:8081
127.0.0.1:8082
127.0.0.1:8083
127.0.0.1:8084
127.0.0.1:8085
```

The first 3 services of the upstream http server should be mapped on the following endpoints via the first envoy proxy:
```
127.0.0.1:9081
127.0.0.1:9082
127.0.0.1:9083
```

And they should also be mapped on the following endpoints via the second envoy proxy:
```
127.0.0.1:10081
127.0.0.1:10082
127.0.0.1:10083
```

You can change the envoy configurations in **load balancer-configs**, run **terraform apply** and watch the envoy proxies automatically adjust to the new configuration through the control plane.

Note that there is an additional upstream server (**upstream-server-2**) that you can start to test health checking adjustments. The second server will bind on **127.0.1.1**. The dns server already answers dns queries with the ips of both upstream servers so you've already been validating the health checking on startup, but it never hurts to validate when the server is up and down.