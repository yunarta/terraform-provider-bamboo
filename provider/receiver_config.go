package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
)

type ConfigurableReceiver interface {
	setConfig(config BambooProviderConfig, client *bamboo.Client)
}

func ConfigureDataSource(receiver ConfigurableReceiver, ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data, ok := request.ProviderData.(*BambooProviderData)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AtlassianCloudProviderModel, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	receiver.setConfig(data.config, data.client)
}

func ConfigureResource(receiver ConfigurableReceiver, ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	data, ok := request.ProviderData.(*BambooProviderData)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AtlassianCloudProviderModel, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	receiver.setConfig(data.config, data.client)
}
