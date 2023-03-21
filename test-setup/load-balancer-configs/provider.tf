provider "etcd" {
  endpoints = "127.0.0.1:32279"
  ca_cert = "${path.module}/../control-plane/etcd_ca.crt"
  cert = "${path.module}/../control-plane/etcd_root.crt"
  key = "${path.module}/../control-plane/etcd_root.key"
}