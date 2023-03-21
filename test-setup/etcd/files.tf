resource "local_file" "etcd_ca_cert" {
  content         = module.etcd.ca_certificate
  file_permission = "0600"
  filename        = "${path.module}/../control-plane/etcd_ca.crt"
}

resource "local_file" "etcd_root_cert" {
  content         = module.etcd.root_certificate
  file_permission = "0600"
  filename        = "${path.module}/../control-plane/etcd_root.crt"
}

resource "local_file" "etcd_root_key" {
  content         = module.etcd.root_key
  file_permission = "0600"
  filename        = "${path.module}/../control-plane/etcd_root.key"
}