---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "mixtool_rules Data Source - terraform-provider-mixtool"
subcategory: ""
description: |-
  Read mixin and render rules as YAML using mixtool
---

# mixtool_rules (Data Source)

Read mixin and render rules as YAML using mixtool

## Example Usage

```terraform
data "mixtool_rules" "example" {
  source = "testdata/mixin.libsonnet"
  jsonnet_path = [
    "testdata/vendor"
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `source` (String) Source Jsonnet file.

### Optional

- `jsonnet_path` (List of String) External variables providing value as a string.

### Read-Only

- `id` (String) sha256 sum of the Jsonnet file
- `rules` (String) Generated Prometheus rules based on the given mixins
