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
var _ datasource.DataSource = &RulesDataSource{}

func NewRulesDataSource() datasource.DataSource {
	return &RulesDataSource{}
}

type RulesDataSource struct{}

func (d *RulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rules"
}

func (d *RulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Read mixin and render rules as YAML using mixtool",

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
			"rules": schema.StringAttribute{
				MarkdownDescription: "Generated Prometheus rules based on the given mixins",
				Computed:            true,
			},
		},
	}
}

func (d *RulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RulesDataSourceModelV0

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

	rules, err := mixer.GenerateRules(source, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while generating rules",
			fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	var i interface{}
	err = json.Unmarshal(rules, &i)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while JSON unmarshalling rules",
			fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	rulesCompiled, err := yaml.Marshal(i)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error while YAML marshalling rules",
			fmt.Sprintf("Original Error: %s", err),
		)
		return
	}

	rulesState := types.StringValue(string(rulesCompiled))
	config.Rules = rulesState
	config.ID = types.StringValue(hash(source))

	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type RulesDataSourceModelV0 struct {
	ID           types.String `tfsdk:"id"`
	Source       types.String `tfsdk:"source"`
	Jsonnet_path types.List   `tfsdk:"jsonnet_path"`
	Rules        types.String `tfsdk:"rules"`
}
