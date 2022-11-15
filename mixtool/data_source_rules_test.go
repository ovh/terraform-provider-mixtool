package mixtool

import (
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMixtoolRules(t *testing.T) {

	renderedTestFile, err := ioutil.ReadFile("./testdata/rules_out/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMixtoolRules,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.mixtool_rules.example", "rules", string(renderedTestFile),
					),
				),
			},
		},
	})
}

const testAccDataSourceMixtoolRules = `
data "mixtool_rules" "example" {
  source = "testdata/mixin.libsonnet"
  jsonnet_path = [
    "testdata/vendor"
  ]
}
`
