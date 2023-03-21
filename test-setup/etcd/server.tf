module "etcd" {
  source = "git::https://github.com/Ferlab-Ste-Justine/terraform-kubernetes-etcd-localhost.git" 
  etcd_nodeport = 32279
}