package provider

import (
	"testing"
)

func TestDeploymentDataSource_Read(t *testing.T) {
	//_ = os.Setenv("TF_ACC", "1")
	//_ = godotenv.Load()
	//
	//transporter := &util.RecordingHttpPayloadTransport{
	//	Transport: transport.NewHttpPayloadTransport(os.Getenv("TF_BAMBOO_ENDPOINT"),
	//		transport.BearerAuthentication{
	//			Token: os.Getenv("TF_BAMBOO_TOKEN"),
	//		},
	//	),
	//}
	//
	//resource.Test(t, resource.TestCase{
	//	ProtoV6ProviderFactories: testAccProvider(transporter),
	//	Steps: []resource.TestStep{
	//		{
	//			Config: `
	//	data "bamboo_deployment" "test" {
	//		name = "application-1-deployment"
	//	}`,
	//			Check: resource.ComposeAggregateTestCheckFunc(
	//				resource.TestCheckResourceAttr("data.bamboo_deployment.test", "id", "6062091"),
	//			),
	//		},
	//	},
	//})
}
