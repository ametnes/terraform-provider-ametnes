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

data "ametnes_location" "location" {
  name = "Ametnes"
  code = "DSL-USE1"
}

data "ametnes_network" "network" {
  name = "NETWORK-EUW7"
  project = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
}
