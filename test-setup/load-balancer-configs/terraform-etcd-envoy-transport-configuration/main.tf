resource "etcd_key" "lb_configs" {
  key = "${var.etcd_prefix}${var.node_id}"
  value = yamlencode(var.load_balancer)
}