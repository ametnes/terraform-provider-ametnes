terraform {
  required_providers {
    ametnes = {
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  token = var.token
  username = var.username
}

resource "ametnes_project" "project" {
  name = "DemoProject"
  description = "DemoProject"
}
