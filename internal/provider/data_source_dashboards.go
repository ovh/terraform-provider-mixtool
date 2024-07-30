package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/monitoring-mixins/mixtool/pkg/mixer"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DashboardsDataSource{}

func NewDashboardsDataSource() datasource.DataSource {
	return &DashboardsDataSource{}
}

type DashboardsDataSource struct{}

func (d *DashboardsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dashboards"
}

func (d *DashboardsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Read mixin and render dashboards as JSON using mixtool",

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
			"dashboards": schema.MapAttribute{
				MarkdownDescription: "",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *DashboardsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DashboardsDataSourceModelV0

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

	dashboards, err := mixer.GenerateDashboards(source, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while generating dashboards",
			fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	rendered := make(map[string]string)

	// Convert dashboardContent from json.RawMessage to string
	for dashboardName, dashboardContent := range dashboards {
		rendered[dashboardName] = strings.TrimSpace(string(dashboardContent))
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	dashboardsState, diags := types.MapValueFrom(ctx, types.StringType, rendered)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = types.StringValue(hash(source))
	config.Dashboards = dashboardsState

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

type DashboardsDataSourceModelV0 struct {
	ID           types.String `tfsdk:"id"`
	Source       types.String `tfsdk:"source"`
	Jsonnet_path types.List   `tfsdk:"jsonnet_path"`
	Dashboards   types.Map    `tfsdk:"dashboards"`
}
