package provider

type AgentAssignmentModel struct {
	//Id             string `tfsdk:"id"`
	Type           string `tfsdk:"type"`
	AgentId        int64  `tfsdk:"agent"`
	ExecutableId   int64  `tfsdk:"executable_id"`
	ExecutableType string `tfsdk:"executable_type"`
}
