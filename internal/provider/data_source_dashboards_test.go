package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDashboardsDataSource(t *testing.T) {

	renderedTestFile, err := os.ReadFile("./testdata/dashboards_out/example.json")
	if err != nil {
		t.Fatal(err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDashboardsDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mixtool_dashboards.example", "dashboards.example.json", string(renderedTestFile)),
				),
			},
		},
	})
}

const testAccDashboardsDataSourceConfig = `
data "mixtool_dashboards" "example" {
	source = "testdata/mixin.libsonnet"
	jsonnet_path = [
	  "testdata/vendor"
	]
  }
`
