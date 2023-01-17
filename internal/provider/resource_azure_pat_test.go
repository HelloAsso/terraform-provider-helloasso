package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAzurePatResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAzurePatResourceConfig(
					AzurePatResourceModel{
						PatName:                 types.StringValue("gitops"),
						AzureDevopsPatScopes:    types.StringValue("vso.code vso.analytics"),
						AppClientID:             types.StringValue("32b1b93f-696e-456a-b632-edc2f4aadcad"),
						Authority:               types.StringValue("https://login.microsoftonline.com/21daa514-c603-41d6-a63a-5add00d26614"),
						AzureDevopsUser:         types.StringValue("patoche@helloasso.com"),
						AzureDevopsPassword:     types.StringValue("patochepass"),
						AzureDevopsPatEndpoint:  types.StringValue("https://vssps.dev.azure.com/helloasso/_apis/tokens/pats"),
						IsAppRegistrationPublic: types.BoolValue(true),
					}, "rotate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "pat_name", "gitops"),
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "app_client_id", "32b1b93f-696e-456a-b632-edc2f4aadcad"),
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "app_client_secret", ""),
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "authority", "https://login.microsoftonline.com/21daa514-c603-41d6-a63a-5add00d26614"),
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "azure_devops_user", "patoche@helloasso.com"),
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "azure_devops_password", "patochepass"),
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "azure_devops_pat_endpoint", "https://vssps.dev.azure.com/helloasso/_apis/tokens/pats"),
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "azure_devops_pat_scopes", "vso.code vso.analytics"),
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "is_app_registration_public", "true"),
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "rotate_when_changed", "rotate"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "helloasso_azure_pat.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{
					"pat_name",
					"app_client_id",
					"app_client_secret",
					"authority",
					"azure_devops_user",
					"azure_devops_password",
					"azure_devops_pat_endpoint",
					"azure_devops_pat_scopes",
					"is_app_registration_public",
					"rotate_when_changed",
				},
			},
			// Update and Read testing
			{
				Config: testAccAzurePatResourceConfig(
					AzurePatResourceModel{
						PatName:                 types.StringValue("gitops"),
						AzureDevopsPatScopes:    types.StringValue("vso.code vso.analytics"),
						AppClientID:             types.StringValue("32b1b93f-696e-456a-b632-edc2f4aadcad"),
						Authority:               types.StringValue("https://login.microsoftonline.com/21daa514-c603-41d6-a63a-5add00d26614"),
						AzureDevopsUser:         types.StringValue("patoche@helloasso.com"),
						AzureDevopsPassword:     types.StringValue("patochepass"),
						AzureDevopsPatEndpoint:  types.StringValue("https://vssps.dev.azure.com/helloasso/_apis/tokens/pats"),
						IsAppRegistrationPublic: types.BoolValue(true),
					},
					"rotate",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("helloasso_azure_pat.test", "configurable_attribute", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccAzurePatResourceConfig(config AzurePatResourceModel, rotation string) string {
	return fmt.Sprintf(`
resource "helloasso_azure_pat" "test" {
	pat_name                   = %q
  azure_devops_pat_scopes    = %q
  app_client_id              = %q
  authority                  = %q
  azure_devops_user          = %q
  azure_devops_password      = %q
  azure_devops_pat_endpoint  = %q
  is_app_registration_public = %t
  rotate_when_changed        = %q

}
`, config.PatName.ValueString(), config.AzureDevopsPatScopes.ValueString(), config.AppClientID.ValueString(), config.Authority.ValueString(), config.AzureDevopsUser.ValueString(), config.AzureDevopsPassword.ValueString(), config.AzureDevopsPatEndpoint, config.IsAppRegistrationPublic.ValueBool(), rotation)
}
