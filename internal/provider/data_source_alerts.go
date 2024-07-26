package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/monitoring-mixins/mixtool/pkg/mixer"
	"gopkg.in/yaml.v2"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AlertsDataSource{}

func NewAlertsDataSource() datasource.DataSource {
	return &AlertsDataSource{}
}

type AlertsDataSource struct{}

func (d *AlertsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alerts"
}

func (d *AlertsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Read mixin and render alerts as YAML using mixtool",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "sha256 sum of the Jsonnet file",
				MarkdownDescription: "sha256 sum of the Jsonnet file",
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "Source Jsonnet file.",
				Required:            true,
			},
			"jsonnet_path": schema.ListAttribute{
				MarkdownDescription: "External variables providing value as a string.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"alerts": schema.StringAttribute{
				MarkdownDescription: "Generated Prometheus alerts based on the given mixins",
				Computed:            true,
			},
		},
	}
}

func (d *AlertsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AlertsDataSourceModelV0

	diags := req.Config.Get(ctx, &config)

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	source := config.Source.ValueString()
	jpath := make([]string, 0)

	diags = config.Jsonnet_path.ElementsAs(ctx, &jpath, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := mixer.GenerateOptions{
		JPaths: jpath,
	}

	alerts, err := mixer.GenerateAlerts(source, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while generating alerts",
			fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	var i interface{}
	err = json.Unmarshal(alerts, &i)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while JSON unmarshalling alerts",
			fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	alertsCompiled, err := yaml.Marshal(i)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while YAML marshalling alerts",
			fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	alertsState := types.StringValue(string(alertsCompiled))
	config.Alerts = alertsState
	config.ID = types.StringValue(hash(source))

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type AlertsDataSourceModelV0 struct {
	ID           types.String `tfsdk:"id"`
	Source       types.String `tfsdk:"source"`
	Jsonnet_path types.List   `tfsdk:"jsonnet_path"`
	Alerts       types.String `tfsdk:"alerts"`
}
