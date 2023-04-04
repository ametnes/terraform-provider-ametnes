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

resource "ametnes_service" "grafana" {
  name = "grafana43333"
  project = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  kind = "grafana:9.3"
  description = "sample grafana"
  network = data.ametnes_network.network.id
  capacity {
    storage = 1
    memory = 1
    cpu = 1
  }

  config = {
    "auth.azuread.client_id" = "SomeText"
    "auth.azuread.client_secret" = "SomeText"
  }
 
  nodes = 1
}

output "gfn_connections" {
  value = ametnes_service.grafana.connections
}