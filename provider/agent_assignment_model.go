package provider

type AgentAssignmentModel struct {
	Id             int64  `tfsdk:"id"`
	Type           string `tfsdk:"type"`
	AgentId        int64  `tfsdk:"agent"`
	ExecutableId   int64  `tfsdk:"executable_id"`
	ExecutableName string `tfsdk:"executable_name"`
	ExecutableType string `tfsdk:"executable_type"`
}
