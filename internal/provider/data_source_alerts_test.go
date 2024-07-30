package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAlertsDataSource(t *testing.T) {

	renderedTestFile, err := os.ReadFile("./testdata/alerts_out/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccAlertsDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.mixtool_alerts.example", "alerts", string(renderedTestFile),
					),
				),
			},
		},
	})
}

const testAccAlertsDataSourceConfig = `
data "mixtool_alerts" "example" {
	source = "testdata/mixin.libsonnet"
	jsonnet_path = [
      "testdata/vendor"
	]
  }
`
