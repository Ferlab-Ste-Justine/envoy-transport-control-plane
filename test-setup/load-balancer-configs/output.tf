data "etcd_prefix_range_end" "envoy" {
  key = "/envoy/"
}

data "etcd_key_range" "envoy" {
  key       = data.etcd_prefix_range_end.envoy.key
  range_end = data.etcd_prefix_range_end.envoy.range_end

  depends_on = [
    module.envoy_one_configs,
    module.envoy_two_configs
  ]
}

output "envoy_one" {
  value = [for elem in data.etcd_key_range.envoy.results: {
    key = elem.key
    create_revision = elem.create_revision
    mod_revision = elem.mod_revision
    version = elem.version
  }]
}