terraform {
  required_providers {
    helloasso = {
      source = "registry.terraform.io/helloasso/helloasso"
    }
    time = {
      source = "registry.terraform.io/hashicorp/time"
    }
  }
}

provider "time" {
}

provider "helloasso" {
}


resource "time_rotating" "rotate_pass" {
  rotation_days = 90
}

resource "helloasso_azure_pat" "example" {
  pat_name                   = "gitops-hap-dev-test-tf"
  azure_devops_pat_scopes    = "vso.code"
  app_client_id              = "32b1b93f-696e-456a-b632-edc2f4aadcad"
  authority                  = "https://login.microsoftonline.com/21daa514-c603-41d6-a63a-5add00d26614"
  azure_devops_user          = "patoche@helloasso.com"
  azure_devops_password      = "PATOCHE PASSWORD"
  azure_devops_pat_endpoint  = "https://vssps.dev.azure.com/helloasso/_apis/tokens/pats"
  is_app_registration_public = true
  rotate_when_changed        = time_rotating.rotate_pass.id
}
