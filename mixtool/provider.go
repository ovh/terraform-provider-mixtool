package mixtool

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

// New returns a func to generate a provider
func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"mixtool_dashboards": dataSourceMixtoolDashboards(),
				"mixtool_rules":      dataSourceMixtoolRules(),
				"mixtool_alerts":     dataSourceMixtoolAlerts(),
			},
			// This provider does not creates resources
			// ResourcesMap: map[string]*schema.Resource{
			// },
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return nil, nil
	}
}
