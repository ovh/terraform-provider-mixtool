output "alerts" {
  value = provider::mixtool::alerts("${path.module}/mixin.libsonnet", ["${path.module}/vendor"])
}
