package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/monitoring-mixins/mixtool/pkg/mixer"
	"gopkg.in/yaml.v2"
)

var _ datasource.DataSource = &mixinDataSource{}

// mixinDataSource renders one output kind of a mixin. The kind is both the
// type name suffix and the name of the computed output attribute.
type mixinDataSource struct {
	kind        string
	description string
	output      schema.Attribute
	render      func(source string, opts mixer.GenerateOptions) (any, error)
}

func NewAlertsDataSource() datasource.DataSource {
	return &mixinDataSource{
		kind:        "alerts",
		description: "Read mixin and render alerts as YAML using mixtool",
		output: schema.StringAttribute{
			MarkdownDescription: "Generated Prometheus alerts based on the given mixins",
			Computed:            true,
		},
		render: func(source string, opts mixer.GenerateOptions) (any, error) {
			return toYAML(mixer.GenerateAlerts(source, opts))
		},
	}
}

func NewRulesDataSource() datasource.DataSource {
	return &mixinDataSource{
		kind:        "rules",
		description: "Read mixin and render rules as YAML using mixtool",
		output: schema.StringAttribute{
			MarkdownDescription: "Generated Prometheus rules based on the given mixins",
			Computed:            true,
		},
		render: func(source string, opts mixer.GenerateOptions) (any, error) {
			return toYAML(mixer.GenerateRules(source, opts))
		},
	}
}

func NewDashboardsDataSource() datasource.DataSource {
	return &mixinDataSource{
		kind:        "dashboards",
		description: "Read mixin and render dashboards as JSON using mixtool",
		output: schema.MapAttribute{
			MarkdownDescription: "",
			Computed:            true,
			ElementType:         types.StringType,
		},
		render: func(source string, opts mixer.GenerateOptions) (any, error) {
			dashboards, err := mixer.GenerateDashboards(source, opts)
			rendered := make(map[string]string, len(dashboards))
			for name, content := range dashboards {
				rendered[name] = strings.TrimSpace(string(content))
			}
			return rendered, err
		},
	}
}

func (d *mixinDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.kind
}

func (d *mixinDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: d.description,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SHA-256 of the rendered output. It changes when the mixin or any file it imports changes.",
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
			d.kind: d.output,
		},
	}
}

func (d *mixinDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var source types.String
	var jsonnetPath types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("source"), &source)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("jsonnet_path"), &jsonnetPath)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jpath []string
	resp.Diagnostics.Append(jsonnetPath.ElementsAs(ctx, &jpath, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rendered, err := d.render(source.ValueString(), mixer.GenerateOptions{JPaths: jpath})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error generating %s", d.kind),
			fmt.Sprintf("generating %s from %s: %s", d.kind, source.ValueString(), err),
		)
		return
	}

	id, err := hash(rendered)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error hashing %s", d.kind),
			fmt.Sprintf("hashing %s from %s: %s", d.kind, source.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source"), source)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("jsonnet_path"), jsonnetPath)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(d.kind), rendered)...)
}

// toYAML keeps the yaml.v2 layout users already have in state; mixer's own
// YAML option indents and wraps differently and would diff every output.
func toYAML(rendered []byte, err error) (any, error) {
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(rendered, &v); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON: %w", err)
	}
	out, err := yaml.Marshal(v)
	return string(out), err
}

// hash hashes string outputs raw, so users can match id against the output.
// json.Marshal sorts map keys, so the dashboards map hashes deterministically.
func hash(rendered any) (string, error) {
	b, ok := rendered.(string)
	raw := []byte(b)
	if !ok {
		var err error
		if raw, err = json.Marshal(rendered); err != nil {
			return "", err
		}
	}
	sha := sha256.Sum256(raw)
	return hex.EncodeToString(sha[:]), nil
}
