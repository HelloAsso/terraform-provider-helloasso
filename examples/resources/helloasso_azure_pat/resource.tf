

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
