data "mixtool_rules" "example" {
  source       = "${path.module}/mixin.libsonnet"
  jsonnet_path = ["${path.module}/vendor"]
}
