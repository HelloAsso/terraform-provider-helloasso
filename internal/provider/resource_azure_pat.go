package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-helloasso/internal/modifiers"

	"os/exec"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	AZ_SCOPE_DEVOPS                    string = "499b84ac-1321-427f-aa17-267ca6975798/.default"
	PAT_API_VERSION                    string = "api-version=7.0-preview.1"
	APP_API_ENDPOINT                   string = "https://graph.microsoft.com/v1.0/applications"
	SWITCH_PRIVATE_PUBLIC_DEFAULT_WAIT int64  = 7
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &AzurePatResource{}
var _ resource.ResourceWithImportState = &AzurePatResource{}

func NewAzurePatResource() resource.Resource {
	return &AzurePatResource{}
}

// AzurePatResource defines the resource implementation.
type AzurePatResource struct {
	client *http.Client
}

// AzurePatResourceModel describes the resource data model.
type AzurePatResourceModel struct {
	PatName                 types.String `tfsdk:"pat_name"`
	AppClientID             types.String `tfsdk:"app_client_id"`
	AppClientSecret         types.String `tfsdk:"app_client_secret"`
	Authority               types.String `tfsdk:"authority"`
	AzureDevopsUser         types.String `tfsdk:"azure_devops_user"`
	AzureDevopsPassword     types.String `tfsdk:"azure_devops_password"`
	AzureDevopsPatEndpoint  types.String `tfsdk:"azure_devops_pat_endpoint"`
	AzureDevopsPatScopes    types.String `tfsdk:"azure_devops_pat_scopes"`
	IsAppRegistrationPublic types.Bool   `tfsdk:"is_app_registration_public"`
	SwitchPrivatePublic     types.Bool   `tfsdk:"az_cli_switch_private_app_public"`
	SwitchPrivatePublicWait types.Int64  `tfsdk:"az_cli_switch_private_app_public_wait_delay"`
	RotateWhenChanged       types.String `tfsdk:"rotate_when_changed"`
	Pat                     types.String `tfsdk:"pat"`
	PatID                   types.String `tfsdk:"pat_id"`
}

type PatToken struct {
	DisplayName     string   `json:"displayName"`
	ValidTo         string   `json:"validTo"`
	Scope           string   `json:"scope"`
	TargetAccounts  []string `json:"targetAccounts"`
	ValidFrom       string   `json:"validFrom"`
	AuthorizationId string   `json:"authorizationId"`
	Token           string   `json:"token"`
}
type PatCreationResponse struct {
	PatToken      PatToken `json:"patToken"`
	PatTokenError string   `json:"patTokenError"`
}

func (r *AzurePatResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_azure_pat"
}

func (r *AzurePatResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example resource",

		Attributes: map[string]schema.Attribute{
			"pat_name": schema.StringAttribute{
				MarkdownDescription: "Name of PAT to create",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"azure_devops_pat_scopes": schema.StringAttribute{
				MarkdownDescription: "Scopes of PAT token separated by a whitespace, see https://github.com/MicrosoftDocs/azure-devops-docs/blob/main/docs/integrate/includes/scopes.md",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_client_id": schema.StringAttribute{
				MarkdownDescription: "Client ID of registered app",
				Required:            true,
			},
			"authority": schema.StringAttribute{
				MarkdownDescription: "AzureAD authority URL",
				Required:            true,
			},
			"azure_devops_user": schema.StringAttribute{
				MarkdownDescription: "Username of Azure Devops user",
				Required:            true,
			},
			"azure_devops_password": schema.StringAttribute{
				MarkdownDescription: "Password of Azure Devops user",
				Required:            true,
				Sensitive:           true,
			},
			"azure_devops_pat_endpoint": schema.StringAttribute{
				MarkdownDescription: "API endpoint to manage PATs",
				Required:            true,
			},
			"is_app_registration_public": schema.BoolAttribute{
				MarkdownDescription: `Is app registration managing authent public or confidential, if false need to set 'app_client_secret'
															Warning WIP: must be 'true' as confidential app is not supported for now`,
				Optional: true,

				PlanModifiers: []planmodifier.Bool{
					modifiers.DefaultBool(true),
				},
			},
			"az_cli_switch_private_app_public": schema.BoolAttribute{
				MarkdownDescription: `This is a dirty workaround to be able to use confidential app with public flow
										to be used with 'is_app_registration_public = true'
										if true, we will call local AZ CLI to switch app public while getting token, after put back to private
										default: false`,
				Optional:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"az_cli_switch_private_app_public_wait_delay": schema.Int64Attribute{
				MarkdownDescription: "When 'az_cli_switch_private_app_public = true' delay to wait for change propagation before acquiring token, (default: 7)",
				Optional:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"app_client_secret": schema.StringAttribute{
				MarkdownDescription: "WIP (not supported): Client secret of registered app (to be set if is_app_registration_public=false)",
				Optional:            true,
				Sensitive:           true,
			},
			"rotate_when_changed": schema.StringAttribute{
				MarkdownDescription: "Arbitrary map of values that, when changed, will trigger rotation of the PAT",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pat": schema.StringAttribute{
				MarkdownDescription: "PAT token",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"pat_id": schema.StringAttribute{
				MarkdownDescription: "PAT ID",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *AzurePatResource) getPublicAdToken(ctx context.Context, appID string, azureUser string, azurePassword string, authority string, apiScope string, switchPrivatePublic bool, switchPrivatePublicWait int64) (string, error) {

	tflog.Info(ctx, "getPublicAdToken")
	// We add a workaround here for more security: make app public only while we get the token
	// Since there is no AcquireTokenByUsernamePassword for Confidential App yet
	azCmd := []string{"az", "ad", "app", "update", "--id", appID, "--is-fallback-public-client"}
	if switchPrivatePublic {
		appPublic := append(azCmd, "true")

		tflog.Info(ctx, "Workaround : Make app public while getting token")
		cmd := exec.Command(appPublic[0], appPublic[1:]...)
		_, err := cmd.Output()

		if err != nil {
			tflog.Info(ctx, err.Error())
		}

		if switchPrivatePublicWait == 0 {
			switchPrivatePublicWait = SWITCH_PRIVATE_PUBLIC_DEFAULT_WAIT
		}
		tflog.Info(ctx, fmt.Sprintf("Workaround : sleep %d sec to take effect", switchPrivatePublicWait))
		time.Sleep(time.Duration(switchPrivatePublicWait) * time.Second)

	}

	// Now get token using public app
	app, err := public.New(appID, public.WithAuthority(authority))

	if err != nil {
		return "", err
	}
	result, err := app.AcquireTokenByUsernamePassword(context.Background(), []string{apiScope}, azureUser, azurePassword)

	if err != nil {
		return "", err
	}

	// We got the token , make it back to private
	if switchPrivatePublic {
		appPrivate := append(azCmd, "false")
		tflog.Info(ctx, "Workaround : Make app back to private now we have the token")
		cmd := exec.Command(appPrivate[0], appPrivate[1:]...)
		_, err = cmd.Output()

		if err != nil {
			tflog.Info(ctx, err.Error())
		}
	}
	return result.AccessToken, nil
}

func (r *AzurePatResource) getConfidentialAdToken(ctx context.Context, appID string, appSecret string, authority string, apiScope string) (string, error) {

	tflog.Info(ctx, "getConfidentialAdToken, first use app secret to connect to app")
	// Initializing the client credential
	cred, err := confidential.NewCredFromSecret(appSecret)
	if err != nil {
		return "", err
	}
	confidentialClientApp, err := confidential.New(authority, appID, cred)

	if err != nil {
		return "", err
	}

	tflog.Info(ctx, "getConfidentialAdToken, second get token")
	result, err := confidentialClientApp.AcquireTokenByCredential(ctx, []string{apiScope})
	if err != nil {
		return "", err
	}

	return result.AccessToken, nil

}

func (r *AzurePatResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *AzurePatResource) deletePat(_ context.Context, patID string, azureDevopsPatEndpoint string, token string) error {

	client := &http.Client{}
	delete_req, _ := http.NewRequest(http.MethodDelete, azureDevopsPatEndpoint+"?"+PAT_API_VERSION+"&authorizationId="+patID, nil)
	delete_req.Header.Set("Authorization", "Bearer "+token)
	res, err := client.Do(delete_req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 && res.StatusCode != 204 {

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("DELETE API did not return 200 or 204 but %d, message %v", res.StatusCode, string(body))
	}
	return nil
}

func (r *AzurePatResource) createPat(_ context.Context, patName string, patScopes string, azureDevopsPatEndpoint string, token string) (*PatCreationResponse, error) {

	now := time.Now()
	expiration := now.AddDate(1, 0, 0)
	postData := map[string]string{
		"allOrgs":     "false",
		"displayName": patName,
		"scope":       patScopes,
		"validTo":     expiration.Format(time.RFC3339),
	}
	json_data, err := json.Marshal(postData)

	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	graph_req, _ := http.NewRequest(http.MethodPost, azureDevopsPatEndpoint+"?"+PAT_API_VERSION, bytes.NewBuffer(json_data))
	graph_req.Header.Set("Authorization", "Bearer "+token)
	graph_req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(graph_req)
	if err != nil {
		return nil, err
	}

	patCreationResponse := &PatCreationResponse{}
	err = json.NewDecoder(res.Body).Decode(patCreationResponse)

	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Create PAT returned %d, error: %v", res.StatusCode, patCreationResponse.PatTokenError)
	}

	return patCreationResponse, nil

}

func (r *AzurePatResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AzurePatResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var accessToken string
	var err error
	if data.IsAppRegistrationPublic.ValueBool() {
		accessToken, err = r.getPublicAdToken(ctx, data.AppClientID.ValueString(), data.AzureDevopsUser.ValueString(), data.AzureDevopsPassword.ValueString(), data.Authority.ValueString(), AZ_SCOPE_DEVOPS, data.SwitchPrivatePublic.ValueBool(), data.SwitchPrivatePublicWait.ValueInt64())
	} else {
		if data.AppClientSecret.IsNull() || data.AppClientSecret.IsUnknown() {
			resp.Diagnostics.AddError("Client Error", "PAT creation: You need to set app_client_secret if is_app_registration_public=false")
			return
		}
		accessToken, err = r.getConfidentialAdToken(ctx, data.AppClientID.ValueString(), data.AppClientSecret.ValueString(), data.Authority.ValueString(), AZ_SCOPE_DEVOPS)
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not get token for PAT creation: %v", err))
		return
	}

	patCreationResponse, err := r.createPat(ctx, data.PatName.ValueString(), data.AzureDevopsPatScopes.ValueString(), data.AzureDevopsPatEndpoint.ValueString(), accessToken)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not get token, check app registration Public status, got error %v", err))
		return
	}

	data.Pat = types.StringValue(patCreationResponse.PatToken.Token)
	data.PatID = types.StringValue(patCreationResponse.PatToken.AuthorizationId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AzurePatResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (r *AzurePatResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AzurePatResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AzurePatResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AzurePatResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.PatID.IsNull() || data.PatID.IsUnknown() {
		tflog.Info(ctx, "Delete state: No PAT ID, nothing to delete")
		return
	} else {

		var accessToken string
		var err error
		if data.IsAppRegistrationPublic.ValueBool() {
			accessToken, err = r.getPublicAdToken(ctx, data.AppClientID.ValueString(), data.AzureDevopsUser.ValueString(), data.AzureDevopsPassword.ValueString(), data.Authority.ValueString(), AZ_SCOPE_DEVOPS, data.SwitchPrivatePublic.ValueBool(), data.SwitchPrivatePublicWait.ValueInt64())
		} else {
			if data.AppClientSecret.IsNull() || data.AppClientSecret.IsUnknown() {
				resp.Diagnostics.AddError("Client Error", "PAT deletion: You need to set app_client_secret if is_app_registration_public=false")
				return
			}
			accessToken, err = r.getConfidentialAdToken(ctx, data.AppClientID.ValueString(), data.AppClientSecret.ValueString(), data.Authority.ValueString(), AZ_SCOPE_DEVOPS)
		}

		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not get token for PAT deletion: %v", err))
			return
		}

		err = r.deletePat(ctx, data.PatID.ValueString(), data.AzureDevopsPatEndpoint.ValueString(), accessToken)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not delete PAT (maybe check app registration public status) err: %v", err))
			return
		}
	}
}

func (r *AzurePatResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
