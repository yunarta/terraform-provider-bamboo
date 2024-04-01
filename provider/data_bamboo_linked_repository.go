package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
	"os"
	"path/filepath"
	"strconv"
)

type LinkedRepositoryData struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var (
	_ datasource.DataSource              = &LinkedRepositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &LinkedRepositoryDataSource{}
	_ ConfigurableReceiver               = &LinkedRepositoryDataSource{}
)

func NewLinkedRepositoryDataSource() datasource.DataSource {
	return &LinkedRepositoryDataSource{}
}

type LinkedRepositoryDataSource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *LinkedRepositoryDataSource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *LinkedRepositoryDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ConfigureDataSource(receiver, ctx, request, response)
}

func (receiver *LinkedRepositoryDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_linked_repository"
}

func (receiver *LinkedRepositoryDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "This data source used define a lookup of linked repository by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed linked repository id.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Linked repository name.",
			},
		},
	}
}

func (receiver *LinkedRepositoryDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		data LinkedRepositoryData
	)

	diags = request.Config.Get(ctx, &data)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	cacheDir := filepath.Join(".cache", "bamboo_linked_repository")
	_ = os.MkdirAll(cacheDir, 0755)
	filename := filepath.Join(cacheDir, fmt.Sprintf("%s.json", data.Name.ValueString()))

	var repository *bamboo.Repository
	if _, err := os.Stat(filename); err == nil {
		fileBytes, err := os.ReadFile(filename)
		if err != nil {
			err = json.Unmarshal(fileBytes, repository)
		}
	}

	if repository == nil {
		repository, err = receiver.client.RepositoryService().Read(data.Name.ValueString())
		if util.TestError(&response.Diagnostics, err, "Failed to retrieve linked repository") {
			return
		}
		marshal, _ := json.Marshal(repository)
		_ = os.WriteFile(filename, marshal, 0644)
	}

	if repository == nil {
		response.Diagnostics.AddError("Missing linked repository", fmt.Sprintf("Unable to find linked repository with name '%s'", data.Name.ValueString()))
		return
	}

	diags = response.State.Set(ctx, &LinkedRepositoryData{
		Id:   types.StringValue(strconv.Itoa(repository.ID)),
		Name: types.StringValue(repository.Name),
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
