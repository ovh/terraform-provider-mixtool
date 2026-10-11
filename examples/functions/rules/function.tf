output "rules" {
  value = provider::mixtool::rules("${path.module}/mixin.libsonnet", ["${path.module}/vendor"])
}
