package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccMixinDataSource_idTracksContent(t *testing.T) {
	source := filepath.Join(t.TempDir(), "mixin.libsonnet")
	writeAlertMixin := func(alert string) {
		mixin := fmt.Sprintf(`{ prometheusAlerts: { groups: [{ name: "g", rules: [{ alert: %q, expr: "1" }] }] } }`, alert)
		if err := os.WriteFile(source, []byte(mixin), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	idChanges := statecheck.CompareValue(compare.ValuesDiffer())

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() { writeAlertMixin("First") },
				Config:    testAccMixinDataSourceConfig_alerts(source),
				Check:     testAccCheckMixinIDIsOutputHash("data.mixtool_alerts.test", "alerts"),
				ConfigStateChecks: []statecheck.StateCheck{
					idChanges.AddStateValue("data.mixtool_alerts.test", tfjsonpath.New("id")),
				},
			},
			{
				PreConfig: func() { writeAlertMixin("Second") },
				Config:    testAccMixinDataSourceConfig_alerts(source),
				ConfigStateChecks: []statecheck.StateCheck{
					idChanges.AddStateValue("data.mixtool_alerts.test", tfjsonpath.New("id")),
				},
			},
		},
	})
}

func testAccCheckMixinIDIsOutputHash(name, output string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attrs := s.RootModule().Resources[name].Primary.Attributes
		sha := sha256.Sum256([]byte(attrs[output]))
		if want := hex.EncodeToString(sha[:]); attrs["id"] != want {
			return fmt.Errorf("id = %s, want sha256 of %s = %s", attrs["id"], output, want)
		}
		return nil
	}
}

func testAccMixinDataSourceConfig_alerts(source string) string {
	return fmt.Sprintf(`
data "mixtool_alerts" "test" {
  source = %[1]q
}
`, source)
}
