package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

func TestAccMixinDataSource_errors(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	cases := map[string]struct{ source, cause string }{
		"missing": {filepath.Join(dir, "missing.libsonnet"), `couldn't\s+open\s+import\s+"[^"]*missing\.libsonnet"`},
		"syntax":  {write("syntax.libsonnet", `{ prometheusAlerts: `), `Unexpected\s+end\s+of\s+file`},
		"import":  {write("import.libsonnet", `(import "nope.libsonnet") + {}`), `couldn't\s+open\s+import\s+"nope\.libsonnet"`},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testAccMixinDataSourceConfig_alerts(tc.source),
						ExpectError: regexp.MustCompile(`(?s)Error generating alerts.*` + tc.cause),
					},
				},
			})
		})
	}
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
