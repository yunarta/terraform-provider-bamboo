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

type ProjectPermissionsData struct {
	Key    string    `tfsdk:"key"`
	Users  types.Map `tfsdk:"users"`
	Groups types.Map `tfsdk:"groups"`
}

var (
	_ datasource.DataSource              = &ProjectPermissionsDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectPermissionsDataSource{}
	_ ConfigurableReceiver               = &ProjectPermissionsDataSource{}
)

func NewProjectPermissionsDataSource() datasource.DataSource {
	return &ProjectPermissionsDataSource{}
}

type ProjectPermissionsDataSource struct {
	client *bamboo.Client
	config BambooProviderConfig
}

func (receiver *ProjectPermissionsDataSource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectPermissionsDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ConfigureDataSource(receiver, ctx, request, response)
}

func (receiver *ProjectPermissionsDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project_permissions"
}

func (receiver *ProjectPermissionsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: `This data source define a lookup of project permissions`,
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project key.",
			},
			"users": schema.MapAttribute{
				Computed: true,
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				MarkdownDescription: "A map with the permission as the key and list of users as the value.",
			},
			"groups": schema.MapAttribute{
				Computed: true,
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				MarkdownDescription: "A map with the permission as the key and list of groups as the value.",
			},
		},
	}
}

func (receiver *ProjectPermissionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		config ProjectPermissionsData
	)

	diags = request.Config.Get(ctx, &config)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	assignedPermissions, err := receiver.client.ProjectService().ReadPermissions(config.Key)
	if util.TestError(&response.Diagnostics, err, "Failed to read deployment repositories") {
		return
	}

	users, groups, diags := CreateAttestation(ctx, assignedPermissions, &response.Diagnostics)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &ProjectPermissionsData{
		Key:    config.Key,
		Users:  users,
		Groups: groups,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
