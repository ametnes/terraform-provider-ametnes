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

data "ametnes_project" "project" {
  name = "Demo"
}

data "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

resource "ametnes_network" "network" {
  name = "NETWORK-EUW5"
  project = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  description = "My loadbalance resource"
}

resource "ametnes_service" "hdb" {
  name = "harperdb43334"
  project = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  kind = "harperdb:3.3"
  description = "sample grafana"
  network = ametnes_network.network.resource_id
  capacity {
    storage = 100
    memory = 4
    cpu = 2
  }

  config = {
    "admin.user" = var.hdb_admin_user
    "admin.password" = var.hdb_admin_pass
    "clustering.user" = var.hdb_clustering_user
    "clustering.password" = var.hdb_clustering_pass
  }
 
  nodes = 1
}

output "hdb_connections" {
  value = ametnes_service.hdb.connections
}
