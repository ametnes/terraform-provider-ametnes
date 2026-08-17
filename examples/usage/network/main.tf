terraform {
  required_providers {
    ametnes = {
      # version = "0.4"
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  token = var.token
  username = var.username
}

# data "ametnes_project" "project" {
#   name = "Demo"
# }

# data "ametnes_location" "location" {
#   name = "Ametnes Cloud"
#   code = "EUW1"
# }

data "ametnes_project" "project" {
  name = "Default"
}

data "ametnes_location" "location" {
  name = "Ametnes"
  code = "DSL-USE1"
}


resource "ametnes_network" "network" {
  name = "NETWORK-EUW8"
  project = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  description = "My loadbalance resource"
}
