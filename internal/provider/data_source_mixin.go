package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/monitoring-mixins/mixtool/pkg/mixer"
	"gopkg.in/yaml.v2"
)

var _ datasource.DataSource = &mixinDataSource{}

type mixinDataSource struct {
	mixinKind
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
				MarkdownDescription: "Path to the mixin Jsonnet file. Terraform resolves a relative path from the working directory, not the module directory, so use `path.module` in modules.",
				Required:            true,
			},
			"jsonnet_path": schema.ListAttribute{
				MarkdownDescription: "Directories to search for Jsonnet imports, like the `-J` flag of `jsonnet` and `mixtool`. Usually the `vendor` directory of `jsonnet-bundler`.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"lint": schema.BoolAttribute{
				MarkdownDescription: "Run `mixtool lint` on the mixin and fail on any finding. `alerts` and `rules` apply the Prometheus guidelines, which check the mixin alerts, so `rules` also fails on alert findings. `dashboards` applies the Grafana dashboard linter. A `.lint` file next to `source` configures exclusions.",
				Optional:            true,
			},
			d.kind: d.output,
		},
	}
}

func (d *mixinDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var source types.String
	var jsonnetPath types.List
	var lint types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("source"), &source)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("jsonnet_path"), &jsonnetPath)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("lint"), &lint)...)
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
		d.addError(resp, "generating", source, err)
		return
	}

	if lint.ValueBool() {
		opts := d.lint
		opts.JPaths = jpath
		var findings strings.Builder
		if err := mixer.Lint(&findings, source.ValueString(), opts); err != nil {
			d.addError(resp, "linting", source, fmt.Errorf("%w\n%s", err, ansiEscape.ReplaceAllString(findings.String(), "")))
			return
		}
	}

	id, err := hash(rendered)
	if err != nil {
		d.addError(resp, "hashing", source, err)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source"), source)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("jsonnet_path"), jsonnetPath)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lint"), lint)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(d.kind), rendered)...)
}

func renderAlerts(source string, opts mixer.GenerateOptions) (any, error) {
	return toYAML(mixer.GenerateAlerts(source, opts))
}

func renderRules(source string, opts mixer.GenerateOptions) (any, error) {
	return toYAML(mixer.GenerateRules(source, opts))
}

func renderDashboards(source string, opts mixer.GenerateOptions) (any, error) {
	dashboards, err := mixer.GenerateDashboards(source, opts)
	rendered := make(map[string]string, len(dashboards))
	for name, content := range dashboards {
		rendered[name] = strings.TrimSpace(string(content))
	}
	return rendered, err
}

// ansiEscape strips the colors mixer.Lint adds when stdout is a terminal, as under go test.
var ansiEscape = regexp.MustCompile("\x1b\\[[0-9;]*m")

func (d *mixinDataSource) addError(resp *datasource.ReadResponse, step string, source types.String, err error) {
	resp.Diagnostics.AddError(
		fmt.Sprintf("Error %s %s", step, d.kind),
		fmt.Sprintf("%s %s from %s: %s", step, d.kind, source.ValueString(), err),
	)
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
