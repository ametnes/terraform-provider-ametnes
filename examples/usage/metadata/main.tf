terraform {
  required_providers {
    ametnes = {
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
   host = "https://cloud.ametnes.com/api/c/v1"
  token = var.token
  username = var.username
}

resource "ametnes_project" "project" {
  name = "DemoProject"
  description = "DemoProject"
}
