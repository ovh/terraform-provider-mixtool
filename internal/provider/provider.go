package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure mixtoolProvider satisfies various provider interfaces.
var _ provider.Provider = &mixtoolProvider{}
var _ provider.ProviderWithFunctions = &mixtoolProvider{}

// mixtoolProvider defines the provider implementation.
type mixtoolProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *mixtoolProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mixtool"
	resp.Version = p.version
}

func (p *mixtoolProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{}
}

func (p *mixtoolProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

}

func (p *mixtoolProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *mixtoolProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	var dataSources []func() datasource.DataSource
	for _, k := range mixinKinds {
		dataSources = append(dataSources, func() datasource.DataSource { return &mixinDataSource{k} })
	}
	return dataSources
}

func (p *mixtoolProvider) Functions(ctx context.Context) []func() function.Function {
	var functions []func() function.Function
	for _, k := range mixinKinds {
		functions = append(functions, func() function.Function { return &mixinFunction{k} })
	}
	return functions
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &mixtoolProvider{
			version: version,
		}
	}
}
