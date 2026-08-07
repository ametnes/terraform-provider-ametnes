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

data "ametnes_project" "project" {
  name = "Default"
}
