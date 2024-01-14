package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
)

type ProjectData struct {
	Key    types.String `tfsdk:"key"`
	Users  types.Map    `tfsdk:"users"`
	Groups types.Map    `tfsdk:"groups"`
}

var (
	_ datasource.DataSource              = &ProjectDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectDataSource{}
	_ ConfigurableReceiver               = &ProjectDataSource{}
)

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

type ProjectDataSource struct {
	client *bamboo.Client
	config BambooProviderConfig
}

func (receiver *ProjectDataSource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ConfigureDataSource(receiver, ctx, request, response)
}

func (receiver *ProjectDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project"
}

func (receiver *ProjectDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required: true,
			},
			"users": schema.MapAttribute{
				Computed: true,
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
			},
			"groups": schema.MapAttribute{
				Computed: true,
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
			},
		},
	}
}

func (receiver *ProjectDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		config ProjectData
		diags  diag.Diagnostics
	)

	diags = request.Config.Get(ctx, &config)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	assignedPermissions, err := receiver.client.ProjectService().ReadPermissions(config.Key.ValueString())
	if util.TestError(&response.Diagnostics, err, "Failed to read deployment repositories") {
		return
	}

	users, groups, diags := CreateAttestation(ctx, assignedPermissions, &response.Diagnostics)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &ProjectData{
		Key:    config.Key,
		Users:  users,
		Groups: groups,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
