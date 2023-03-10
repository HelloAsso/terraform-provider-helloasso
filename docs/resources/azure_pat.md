---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "helloasso_azure_pat Resource - terraform-provider-helloasso"
subcategory: ""
description: |-
  Example resource
---

# helloasso_azure_pat (Resource)

Example resource

## Example Usage

```terraform
resource "time_rotating" "rotate_pass" {
  rotation_days = 90
}

resource "helloasso_azure_pat" "example" {
  pat_name                  = "gitops"
  azure_devops_pat_scopes   = "vso.code"
  app_client_id             = "6145d7e0-7adf-4a48-b516-4f61cb047efd"
  authority                 = "https://login.microsoftonline.com/128c6ba1-f30f-4176-87d4-a93c61ae4ef0"
  azure_devops_user         = "user@myorganization.com"
  azure_devops_password     = "usersuperpassword"
  azure_devops_pat_endpoint = "https://vssps.dev.azure.com/myorganization/_apis/tokens/pats"

  # For now it does not support private App - must be true
  is_app_registration_public = true

  rotate_when_changed = time_rotating.rotate_pass.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_client_id` (String) Client ID of registered app
- `authority` (String) AzureAD authority URL
- `azure_devops_password` (String, Sensitive) Password of Azure Devops user
- `azure_devops_pat_endpoint` (String) API endpoint to manage PATs
- `azure_devops_pat_scopes` (String) Scopes of PAT token separated by a whitespace, see https://github.com/MicrosoftDocs/azure-devops-docs/blob/main/docs/integrate/includes/scopes.md
- `azure_devops_user` (String) Username of Azure Devops user
- `pat_name` (String) Name of PAT to create

### Optional

- `app_client_secret` (String, Sensitive) WIP (not supported): Client secret of registered app (to be set if is_app_registration_public=false)
- `is_app_registration_public` (Boolean) Is app registration managing authent public or confidential, if false need to set 'app_client_secret'
															Warning WIP: must be 'true' as confidential app is not supported for now
- `rotate_when_changed` (String) Arbitrary map of values that, when changed, will trigger rotation of the PAT

### Read-Only

- `pat` (String, Sensitive) PAT token
- `pat_id` (String) PAT ID


