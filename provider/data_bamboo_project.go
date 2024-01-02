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
	Users  types.List   `tfsdk:"users"`
	Groups types.List   `tfsdk:"groups"`
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
			"users": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"groups": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (receiver *ProjectDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		data  ProjectData
		diags diag.Diagnostics
	)

	diags = request.Config.Get(ctx, &data)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	permissions, err := receiver.client.ProjectService().ReadPermissions(data.Key.ValueString())
	if util.TestError(&response.Diagnostics, err, "Failed to read deployment repositories") {
		return
	}

	var (
		users  = make([]string, 0)
		groups = make([]string, 0)
	)
	for _, user := range permissions.Users {
		users = append(users, user.Name)
	}

	for _, group := range permissions.Groups {
		groups = append(groups, group.Name)
	}

	userList, diags := types.ListValueFrom(ctx, types.StringType, users)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	groupList, diags := types.ListValueFrom(ctx, types.StringType, groups)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &ProjectData{
		Key:    types.StringValue(data.Key.ValueString()),
		Users:  userList,
		Groups: groupList,
	})

	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
