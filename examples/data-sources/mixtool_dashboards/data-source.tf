data "mixtool_dashboards" "example" {
  source       = "${path.module}/mixin.libsonnet"
  jsonnet_path = ["${path.module}/vendor"]
}
