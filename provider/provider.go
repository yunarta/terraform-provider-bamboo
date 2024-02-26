package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yunarta/terraform-api-transport/transport"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"os"
	"path/filepath"
)

type BambooProvider struct {
	Version string
}

func (p *BambooProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "bamboo"
	response.Version = p.Version
}

func (p *BambooProvider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Blocks: map[string]schema.Block{
			"bamboo": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Required: true,
					},
					"token": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
				},
			},
			"bamboo_rss": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"server": schema.StringAttribute{
						Required: true,
					},
					"name": schema.StringAttribute{
						Required: true,
					},
					"clone_url": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
	}
}

func (p *BambooProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config BambooProviderConfig

	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	_ = os.RemoveAll(filepath.Join(".cache"))

	providerData := &BambooProviderData{
		config: config,
		client: bamboo.NewBambooClient(
			transport.NewHttpPayloadTransport(config.Bamboo.EndPoint.ValueString(),
				transport.BearerAuthentication{
					Token: config.Bamboo.Token.ValueString(),
				},
			),
		),
	}

	response.DataSourceData = providerData
	response.ResourceData = providerData
}

func (p *BambooProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLinkedRepositoryDataSource,
		NewDeploymentDataSource,
		NewProjectDataSource,
	}
}

func (p *BambooProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAgentAssignmentResource,
		NewProjectResource,
		NewProjectVariableResource,
		NewProjectPermissionsResource,
		NewProjectRepositoriesResource,
		NewDeploymentResource,
		NewDeploymentRepositoryResource,
		NewProjectLinkedRepositoryResource,
		NewLinkedRepositoryResource,
		NewLinkedRepositoryAccessorResource,
		NewLinkedRepositoryDependencyResource,
	}
}

var _ provider.Provider = &BambooProvider{}

func New(Version string) func() provider.Provider {
	return func() provider.Provider {
		return &BambooProvider{
			Version: Version,
		}
	}
}
