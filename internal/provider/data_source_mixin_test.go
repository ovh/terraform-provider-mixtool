package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

type mixinKindTest struct {
	kind   string
	golden knownvalue.Check
	// id is computed outside the provider: a change here changes every user's id.
	id       string
	lintRule string
}

// Golden files are the expected output; never regenerate them from this provider.
func mixinKindTests(t *testing.T) []mixinKindTest {
	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	sha := func(s string) string {
		sum := sha256.Sum256([]byte(s))
		return hex.EncodeToString(sum[:])
	}
	alerts, rules := read("alerts_out/example.yaml"), read("rules_out/example.yaml")

	return []mixinKindTest{
		{
			kind:     "alerts",
			golden:   knownvalue.StringExact(alerts),
			id:       sha(alerts),
			lintRule: `alert-summary-missing-rule`,
		},
		{
			kind:     "rules",
			golden:   knownvalue.StringExact(rules),
			id:       sha(rules),
			lintRule: `alert-summary-missing-rule`,
		},
		{
			kind: "dashboards",
			golden: knownvalue.MapExact(map[string]knownvalue.Check{
				"example.json": knownvalue.StringExact(strings.TrimSpace(read("dashboards_out/example.json"))),
				"second.json":  knownvalue.StringExact(strings.TrimSpace(read("dashboards_out/second.json"))),
			}),
			id:       "1607784f4a7872360fd1dc1b791e09b711caa6758df7778c48852a6fca999e39",
			lintRule: `panel-units-rule`,
		},
	}
}

func TestMixinDataSource_kinds(t *testing.T) {
	for _, tc := range mixinKindTests(t) {
		t.Run(tc.kind, func(t *testing.T) {
			address := "data.mixtool_" + tc.kind + ".test"
			config := func(lint bool) string {
				return fmt.Sprintf(`
data "mixtool_%[1]s" "test" {
  source       = "testdata/mixin.libsonnet"
  jsonnet_path = ["testdata/vendor"]
  lint         = %[2]t
}
`, tc.kind, lint)
			}

			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: config(false),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(address, tfjsonpath.New(tc.kind), tc.golden),
							statecheck.ExpectKnownValue(address, tfjsonpath.New("id"), knownvalue.StringExact(tc.id)),
						},
					},
					{
						Config:      config(true),
						ExpectError: regexp.MustCompile(`(?s)Error linting ` + tc.kind + `.*\[` + tc.lintRule + `\]`),
					},
				},
			})
		})
	}
}

func TestMixinFunction_kinds(t *testing.T) {
	for _, tc := range mixinKindTests(t) {
		t.Run(tc.kind, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`
output "rendered" {
  value = provider::mixtool::%[1]s("testdata/mixin.libsonnet", ["testdata/vendor"])
}
`, tc.kind),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownOutputValue("rendered", tc.golden),
						},
					},
				},
			})
		})
	}
}

func TestMixinDataSource_idTracksContent(t *testing.T) {
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

func writeFile(t *testing.T, dir, name, content string) string {
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

type renderErrorCase struct{ source, cause string }

func renderErrorCases(t *testing.T) map[string]renderErrorCase {
	dir := t.TempDir()
	return map[string]renderErrorCase{
		"missing": {filepath.Join(dir, "missing.libsonnet"), `couldn't\s+open\s+import\s+"[^"]*missing\.libsonnet"`},
		"syntax":  {writeFile(t, dir, "syntax.libsonnet", `{ prometheusAlerts: `), `Unexpected\s+end\s+of\s+file`},
		"import":  {writeFile(t, dir, "import.libsonnet", `(import "nope.libsonnet") + {}`), `couldn't\s+open\s+import\s+"nope\.libsonnet"`},
	}
}

func TestMixinDataSource_errors(t *testing.T) {
	for _, k := range mixinKindTests(t) {
		for name, tc := range renderErrorCases(t) {
			t.Run(k.kind+"/"+name, func(t *testing.T) {
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: fmt.Sprintf(`
data "mixtool_%[1]s" "test" {
  source = %[2]q
}
`, k.kind, tc.source),
							ExpectError: regexp.MustCompile(`(?s)Error generating ` + k.kind + `.*` + tc.cause),
						},
					},
				})
			})
		}
	}
}

func TestMixinDataSource_emptySource(t *testing.T) {
	for _, k := range mixinKindTests(t) {
		t.Run(k.kind, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`
data "mixtool_%[1]s" "test" {
  source = ""
}
`, k.kind),
						ExpectError: regexp.MustCompile(`(?s)Invalid source.*source must not be empty`),
					},
				},
			})
		})
	}
}

func TestMixinFunction_errors(t *testing.T) {
	for _, k := range mixinKindTests(t) {
		for name, tc := range renderErrorCases(t) {
			t.Run(k.kind+"/"+name, func(t *testing.T) {
				resource.UnitTest(t, resource.TestCase{
					TerraformVersionChecks:   []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_8_0)},
					ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: fmt.Sprintf(`
output "rendered" {
  value = provider::mixtool::%[1]s(%[2]q, [])
}
`, k.kind, tc.source),
							ExpectError: regexp.MustCompile(`(?s)Invalid\s+value\s+for\s+"source"\s+parameter.*generating\s+` + k.kind + `.*` + tc.cause),
						},
					},
				})
			})
		}
	}
}

func TestMixinDataSource_lint(t *testing.T) {
	dir := t.TempDir()
	// Lint-clean except the alert name, so the .lint exclusion is the only thing that can make it pass.
	source := writeFile(t, dir, "mixin.libsonnet", `{ prometheusAlerts: { groups: [{ name: "g", rules: [{
  alert: "not_camel_case",
  expr: "1",
  labels: { severity: "critical" },
  annotations: { description: "{{ $labels.job }} is down.", summary: "Job is down." },
}] }] } }`)
	broken := writeFile(t, dir, "broken.libsonnet", `{ prometheusAlerts: `)
	lintConfig := func(source string) string {
		return fmt.Sprintf(`
data "mixtool_alerts" "test" {
  source = %[1]q
  lint   = true
}
`, source)
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      lintConfig(source),
				ExpectError: regexp.MustCompile(`(?s)Error linting alerts.*\[alert-name-camelcase\]`),
			},
			{
				PreConfig: func() {
					writeFile(t, dir, ".lint", "exclusions:\n  alert-name-camelcase:\n    entries:\n      - alert: not_camel_case\n")
				},
				Config: lintConfig(source),
			},
			{
				Config:      lintConfig(broken),
				ExpectError: regexp.MustCompile(`Error generating alerts`),
			},
		},
	})
}

func TestHash(t *testing.T) {
	for name, tc := range map[string]struct {
		rendered any
		want     string
	}{
		"string is hashed raw": {"groups: []\n", "761adf8d97e15214e4da44effe63bc331092f597e3089818d5825c87444b1f27"},
		"map keys are sorted":  {map[string]string{"b.json": "{}", "a.json": "{}"}, "a40955bdfa956715e11877db1fa9c7a302fb39b9871f40a9902aeeed2f43f81c"},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := hash(tc.rendered)
			if err != nil || got != tc.want {
				t.Fatalf("hash() = %q, %v, want %q", got, err, tc.want)
			}
		})
	}
}

func TestToYAML(t *testing.T) {
	got, err := toYAML([]byte(`{"groups":[{"name":"g","rules":[{"alert":"A","expr":"up == 0"}]}]}`), nil)
	want := "groups:\n- name: g\n  rules:\n  - alert: A\n    expr: up == 0\n"
	if err != nil || got != want {
		t.Fatalf("toYAML() = %q, %v, want %q", got, err, want)
	}

	if _, err := toYAML(nil, fmt.Errorf("render failed")); err == nil || err.Error() != "render failed" {
		t.Fatalf("toYAML() did not pass the render error through: %v", err)
	}
	if _, err := toYAML([]byte(`{`), nil); err == nil {
		t.Fatal("toYAML() accepted invalid JSON")
	}
}

func testAccMixinDataSourceConfig_alerts(source string) string {
	return fmt.Sprintf(`
data "mixtool_alerts" "test" {
  source = %[1]q
}
`, source)
}

func TestANSIEscape(t *testing.T) {
	if got := ansiEscape.ReplaceAllString("\x1b[31m[alert-name-camelcase] bad\x1b[0m", ""); got != "[alert-name-camelcase] bad" {
		t.Fatalf("ansiEscape left %q", got)
	}
}
