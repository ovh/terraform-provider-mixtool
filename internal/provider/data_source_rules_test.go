package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRulesDataSource(t *testing.T) {

	renderedTestFile, err := os.ReadFile("./testdata/rules_out/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccRulesDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.mixtool_rules.example", "rules", string(renderedTestFile),
					),
				),
			},
		},
	})
}

const testAccRulesDataSourceConfig = `
data "mixtool_rules" "example" {
	source = "testdata/mixin.libsonnet"
	jsonnet_path = [
      "testdata/vendor"
	]
  }
`
