package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/monitoring-mixins/mixtool/pkg/mixer"
)

// mixinKind is one output of a mixin. Each kind is served as a data source and
// a function, both named after kind, sharing render; only the data source lints.
type mixinKind struct {
	kind        string
	description string
	output      schema.Attribute
	summary     string
	result      function.Return
	lint        mixer.LintOptions
	render      func(source string, opts mixer.GenerateOptions) (any, error)
}

var mixinKinds = []mixinKind{
	{
		kind:        "alerts",
		description: "Read mixin and render alerts as YAML using mixtool",
		output: schema.StringAttribute{
			MarkdownDescription: "Generated Prometheus alerts based on the given mixins",
			Computed:            true,
		},
		summary: "Render the Prometheus alerts of a mixin as YAML",
		result:  function.StringReturn{},
		lint:    mixer.LintOptions{Prometheus: true},
		render:  renderAlerts,
	},
	{
		kind:        "rules",
		description: "Read mixin and render rules as YAML using mixtool",
		output: schema.StringAttribute{
			MarkdownDescription: "Generated Prometheus rules based on the given mixins",
			Computed:            true,
		},
		summary: "Render the Prometheus rules of a mixin as YAML",
		result:  function.StringReturn{},
		lint:    mixer.LintOptions{Prometheus: true},
		render:  renderRules,
	},
	{
		kind:        "dashboards",
		description: "Read mixin and render dashboards as JSON using mixtool",
		output: schema.MapAttribute{
			MarkdownDescription: "Generated Grafana dashboards as JSON strings, keyed by dashboard file name.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		summary: "Render the Grafana dashboards of a mixin as JSON, keyed by file name",
		result:  function.MapReturn{ElementType: types.StringType},
		lint:    mixer.LintOptions{Grafana: true},
		render:  renderDashboards,
	},
}
