package mixtool

import (
	"io/ioutil"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceMixtoolAlerts(t *testing.T) {

	renderedTestFile, err := ioutil.ReadFile("./testdata/alerts_out/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMixtoolAlerts,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.mixtool_alerts.example", "alerts", string(renderedTestFile),
					),
				),
			},
		},
	})
}

const testAccDataSourceMixtoolAlerts = `
data "mixtool_alerts" "example" {
  source = "testdata/mixin.libsonnet"
  jsonnet_path = [
    "testdata/vendor"
  ]
}
`
