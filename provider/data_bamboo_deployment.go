package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
	"strconv"
)

type DeploymentData struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Users  types.Map    `tfsdk:"users"`
	Groups types.Map    `tfsdk:"groups"`
}

var (
	_ datasource.DataSource              = &DeploymentDataSource{}
	_ datasource.DataSourceWithConfigure = &DeploymentDataSource{}
	_ ConfigurableReceiver               = &DeploymentDataSource{}
)

func NewDeploymentDataSource() datasource.DataSource {
	return &DeploymentDataSource{}
}

type DeploymentDataSource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *DeploymentDataSource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *DeploymentDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	ConfigureDataSource(receiver, ctx, request, response)
}

func (receiver *DeploymentDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment"
}

func (receiver *DeploymentDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
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

func (receiver *DeploymentDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var (
		data  DeploymentData
		diags diag.Diagnostics
	)

	diags = request.Config.Get(ctx, &data)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deployment, err := receiver.client.DeploymentService().Read(data.Name.ValueString())
	if util.TestError(&response.Diagnostics, err, errorFailedToReadDeployment) {
		return
	}

	if deployment == nil {
		response.Diagnostics.AddError("Unable to find deployment", data.Name.ValueString())
		return
	}

	assignedPermissions, err := receiver.client.DeploymentService().ReadPermissions(deployment.ID)
	if util.TestError(&response.Diagnostics, err, "Failed to read deployment permissions") {
		return
	}

	users, groups, diags := CreateAttestation(ctx, assignedPermissions, &response.Diagnostics)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &DeploymentData{
		Id:     types.StringValue(strconv.Itoa(deployment.ID)),
		Name:   types.StringValue(deployment.Name),
		Users:  users,
		Groups: groups,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
