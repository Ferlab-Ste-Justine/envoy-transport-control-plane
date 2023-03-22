variable "etcd_prefix" {
  description = "Prefix of the keyspace to work on"
  type = string
}

variable "node_id" {
  description = "Envoy node id the configuration is for"
  type = string
}

variable "load_balancer" {
  description = "Load balancer configs"
  type = object({
    services = list(object({
      name              = string
      listening_ip      = string
      listening_port    = number
      cluster_domain    = string
      cluster_port      = number
	  idle_timeout      = string
      max_connections   = number
      access_log_format = string
      health_check      = object({
        timeout             = string
        interval            = string
        healthy_threshold   = number
        unhealthy_threshold = number
      })
    }))
    dns_servers = list(object({
      ip   = string
      port = number
    }))
  })
}