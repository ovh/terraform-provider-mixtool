package mixtool

import (
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMixtoolDashboards(t *testing.T) {

	renderedTestFile, err := ioutil.ReadFile("./testdata/dashboards_out/example.json")
	if err != nil {
		t.Fatal(err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMixtoolDashboards,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.mixtool_dashboards.example", "dashboards.example.json", string(renderedTestFile),
					),
				),
			},
		},
	})
}

const testAccDataSourceMixtoolDashboards = `
data "mixtool_dashboards" "example" {
  source = "testdata/mixin.libsonnet"
  jsonnet_path = [
    "testdata/vendor"
  ]
}
`
