resource "kcli_vm" "bootstrap-vm" {
  name = "${var.cluster_id}-bootstrap"
  image = var.image
  overrides = "{'memory': 6144, 'numcpus': 4, 'nets': [\"var.external_bridge\",\"var.provisioning_bridge\"]}"
  ignition = "${var.cluster_id}-bootstrap.ign}"
}
