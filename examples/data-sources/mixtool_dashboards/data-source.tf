data "mixtool_dashboards" "example" {
  source = "testdata/mixin.libsonnet"
  jsonnet_path = [
    "testdata/vendor"
  ]
}
