module "envoy_one_configs" {
  source = "git::https://github.com/Ferlab-Ste-Justine/terraform-etcd-envoy-transport-configuration.git"
  etcd_prefix = "/envoy/"
  node_id = "envoy-one"
  load_balancer = {
    dns_servers = [{
      ip   = "127.0.0.1"
      port = 1053 
    }]
    services = [
      {
        name              = "server1"
        listening_port    = 9081
        listening_ip      = "127.0.0.1"
        cluster_domain    = "test.local"
        cluster_port      = 8081
	    idle_timeout      = "10s"
        max_connections   = 100
        access_log_format = "[%START_TIME%][Connection] %DOWNSTREAM_REMOTE_ADDRESS% to %UPSTREAM_HOST% (cluster %UPSTREAM_CLUSTER%). Error flags: %RESPONSE_FLAGS%\n"
        health_check      = {
          timeout             = "5s"
          interval            = "5s"
          healthy_threshold   = 1
          unhealthy_threshold = 2
        }
        tls_termination   = {
          listener_certificate = ""
          listener_key         = ""
        }
      },
      {
        name              = "server2"
        listening_port    = 9082
        listening_ip      = "127.0.0.1"
        cluster_domain    = "test.local"
        cluster_port      = 8082
	    idle_timeout      = "10s"
        max_connections   = 100
        access_log_format = "[%START_TIME%][Connection] %DOWNSTREAM_REMOTE_ADDRESS% to %UPSTREAM_HOST% (cluster %UPSTREAM_CLUSTER%). Error flags: %RESPONSE_FLAGS%\n"
        health_check      = {
          timeout             = "5s"
          interval            = "5s"
          healthy_threshold   = 1
          unhealthy_threshold = 2
        }
        tls_termination   = {
          listener_certificate = ""
          listener_key         = ""
        }
      },
      {
        name              = "server3"
        listening_port    = 9083
        listening_ip      = "127.0.0.1"
        cluster_domain    = "test.local"
        cluster_port      = 8083
	    idle_timeout      = "10s"
        max_connections   = 100
        access_log_format = "[%START_TIME%][Connection] %DOWNSTREAM_REMOTE_ADDRESS% to %UPSTREAM_HOST% (cluster %UPSTREAM_CLUSTER%). Error flags: %RESPONSE_FLAGS%\n"
        health_check      = {
          timeout             = "5s"
          interval            = "5s"
          healthy_threshold   = 1
          unhealthy_threshold = 2
        }
        tls_termination   = {
          listener_certificate = ""
          listener_key         = ""
        }
      }
    ]
  }
}

module "envoy_two_configs" {
  source = "git::https://github.com/Ferlab-Ste-Justine/terraform-etcd-envoy-transport-configuration.git"
  etcd_prefix = "/envoy/"
  node_id = "envoy-two"
  load_balancer = {
    dns_servers = [{
      ip   = "127.0.0.1"
      port = 1053 
    }]
    services = [
      {
        name            = "server1"
        listening_port  = 10081
        listening_ip    = "127.0.0.1"
        cluster_domain  = "test.local"
        cluster_port    = 8081
	    idle_timeout    = "10s"
        max_connections = 100
        access_log_format = "[%START_TIME%][Connection] %DOWNSTREAM_REMOTE_ADDRESS% to %UPSTREAM_HOST% (cluster %UPSTREAM_CLUSTER%). Error flags: %RESPONSE_FLAGS%\n"
        health_check    = {
          timeout             = "5s"
          interval            = "5s"
          healthy_threshold   = 1
          unhealthy_threshold = 2
        }
        tls_termination   = {
          listener_certificate = ""
          listener_key         = ""
        }
      },
      {
        name            = "server2"
        listening_port  = 10082
        listening_ip    = "127.0.0.1"
        cluster_domain  = "test.local"
        cluster_port    = 8082
	    idle_timeout    = "10s"
        max_connections = 100
        access_log_format = "[%START_TIME%][Connection] %DOWNSTREAM_REMOTE_ADDRESS% to %UPSTREAM_HOST% (cluster %UPSTREAM_CLUSTER%). Error flags: %RESPONSE_FLAGS%\n"
        health_check    = {
          timeout             = "5s"
          interval            = "5s"
          healthy_threshold   = 1
          unhealthy_threshold = 2
        }
        tls_termination   = {
          listener_certificate = ""
          listener_key         = ""
        }
      },
      {
        name            = "server3"
        listening_port  = 10083
        listening_ip    = "127.0.0.1"
        cluster_domain  = "test.local"
        cluster_port    = 8083
	    idle_timeout    = "10s"
        max_connections = 100
        access_log_format = "[%START_TIME%][Connection] %DOWNSTREAM_REMOTE_ADDRESS% to %UPSTREAM_HOST% (cluster %UPSTREAM_CLUSTER%). Error flags: %RESPONSE_FLAGS%\n"
        health_check    = {
          timeout             = "5s"
          interval            = "5s"
          healthy_threshold   = 1
          unhealthy_threshold = 2
        }
        tls_termination   = {
          listener_certificate = ""
          listener_key         = ""
        }
      }
    ]
  }
}