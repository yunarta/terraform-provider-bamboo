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
		MarkdownDescription: `Bamboo provider.
`,
		Blocks: map[string]schema.Block{
			"bamboo": schema.SingleNestedBlock{
				MarkdownDescription: `Bamboo integration definition.`,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: `Bamboo end point url without trailing slash.`,
					},
					"token": schema.StringAttribute{
						Required:            true,
						Sensitive:           true,
						MarkdownDescription: `Bamboo personal access token.`,
					},
				},
			},
			"bamboo_rss": schema.SingleNestedBlock{
				MarkdownDescription: `Bamboo RSS definition.

In order to get the value properly, you need to export your linked repository into local file system and retrieve the value from the exported YAML.

See bamboo rest/api/1.0/export/repository/name/{name}. 

Export configuration of a linked repository to YAML format
`,
				Attributes: map[string]schema.Attribute{
					"server": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: `Linked Bitbucket data center UUID for linked repository and Bamboo Spec management.`,
					},
					"name": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: `Linked Bitbucket data center name`,
					},
					"clone_url": schema.StringAttribute{
						Required: true,
						MarkdownDescription: `Clone URL of the Bitbucket data center.

Must be in following format ssh://git@[bitbucket-hostname]:[bitbucket-ssh-port-number]/%s/%s.git.

Example ssh://git@bitbucket.mobilesolutionworks.com:7999/%s/%s.git`,
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
		NewProjectPermissionsDataSource,
	}
}

func (p *BambooProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPlanResource,
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
