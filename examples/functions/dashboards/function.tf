output "dashboards" {
  value = provider::mixtool::dashboards("${path.module}/mixin.libsonnet", ["${path.module}/vendor"])
}
