package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/yunarta/terraform-api-transport/transport"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"os"
)

func testAccProvider(transport transport.PayloadTransport) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"bamboo": providerserver.NewProtocol6WithError(&RecordingBambooProvider{
			client: bamboo.NewBambooClient(transport),
		}),
	}
}

type RecordingBambooProvider struct {
	provider BambooProvider
	client   *bamboo.Client
}

func (p *RecordingBambooProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	p.provider.Metadata(ctx, request, response)
}

func (p *RecordingBambooProvider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {

}

func (p *RecordingBambooProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {

	providerData := &BambooProviderData{
		config: BambooProviderConfig{
			Bamboo: EndPoint{},
			BambooRss: BambooRss{
				Server:   types.StringValue(os.Getenv("TF_BAMBOORSS_SERVER")),
				Name:     types.StringValue(os.Getenv("TF_BAMBOORSS_NAME")),
				CloneUrl: types.StringValue(os.Getenv("TF_BAMBOORSS_CLONEURL")),
			},
		},
		client: p.client,
	}

	response.DataSourceData = providerData
	response.ResourceData = providerData
}

func (p *RecordingBambooProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return p.provider.DataSources(ctx)
}

func (p *RecordingBambooProvider) Resources(ctx context.Context) []func() resource.Resource {
	return p.provider.Resources(ctx)
}

var _ provider.Provider = &RecordingBambooProvider{}
