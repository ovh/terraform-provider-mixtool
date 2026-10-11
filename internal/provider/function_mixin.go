package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/monitoring-mixins/mixtool/pkg/mixer"
)

var _ function.Function = &mixinFunction{}

type mixinFunction struct {
	mixinKind
}

func (f *mixinFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.kind
}

func (f *mixinFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:             f.summary,
		MarkdownDescription: fmt.Sprintf("Renders the same output as the `mixtool_%s` data source, without a data block. Requires Terraform 1.8 or later.", f.kind),
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "source",
				MarkdownDescription: "Path to the mixin Jsonnet file. Use `path.module` in modules.",
			},
			function.ListParameter{
				Name:                "jsonnet_path",
				MarkdownDescription: "Directories to search for Jsonnet imports. Pass `[]` if the mixin has no imports.",
				ElementType:         types.StringType,
			},
		},
		Return: f.result,
	}
}

func (f *mixinFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var source string
	var jpath []string
	resp.Error = req.Arguments.Get(ctx, &source, &jpath)
	if resp.Error != nil {
		return
	}

	rendered, err := f.render(source, mixer.GenerateOptions{JPaths: jpath})
	if err != nil {
		resp.Error = function.NewArgumentFuncError(0, fmt.Sprintf("generating %s from %s: %s", f.kind, source, err))
		return
	}

	resp.Error = resp.Result.Set(ctx, rendered)
}
