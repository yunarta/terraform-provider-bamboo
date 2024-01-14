package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
)

type EndPoint struct {
	EndPoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

type BambooRss struct {
	Server   types.String `tfsdk:"server"`
	Name     types.String `tfsdk:"name"`
	CloneUrl types.String `tfsdk:"clone_url"`
}

type BambooProviderConfig struct {
	Bamboo    EndPoint  `tfsdk:"bamboo"`
	BambooRss BambooRss `tfsdk:"bamboo_rss"`
}

type BambooProviderData struct {
	config BambooProviderConfig
	client *bamboo.Client
}
