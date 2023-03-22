terraform {
  required_providers {
    ametnes = {
      # version = "0.3"
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  // add you provider here
   host = "https://cloud.ametnes.com/api/c/v1"
  token = var.token
  username = var.username
}

resource "ametnes_project" "project" {
  name = "DemoProject"
  description = "DemoProject"
}
