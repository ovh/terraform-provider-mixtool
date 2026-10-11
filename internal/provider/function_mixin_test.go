package provider

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccMixinFunctions(t *testing.T) {
	read := func(name string) string {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
locals {
  source       = "testdata/mixin.libsonnet"
  jsonnet_path = ["testdata/vendor"]
}

output "alerts" {
  value = provider::mixtool::alerts(local.source, local.jsonnet_path)
}

output "rules" {
  value = provider::mixtool::rules(local.source, local.jsonnet_path)
}

output "dashboards" {
  value = provider::mixtool::dashboards(local.source, local.jsonnet_path)
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("alerts", knownvalue.StringExact(read("testdata/alerts_out/example.yaml"))),
					statecheck.ExpectKnownOutputValue("rules", knownvalue.StringExact(read("testdata/rules_out/example.yaml"))),
					statecheck.ExpectKnownOutputValue("dashboards", knownvalue.MapExact(map[string]knownvalue.Check{
						"example.json": knownvalue.StringExact(strings.TrimSpace(read("testdata/dashboards_out/example.json"))),
					})),
				},
			},
		},
	})
}
