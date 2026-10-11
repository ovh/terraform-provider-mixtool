data "mixtool_alerts" "example" {
  source       = "${path.module}/mixin.libsonnet"
  jsonnet_path = ["${path.module}/vendor"]
}
