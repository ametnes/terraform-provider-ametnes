terraform {
  required_providers {
    ametnes = {
      source  = "ametnes.com/cloud/ametnes"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

provider "ametnes" {
  token    = var.token
  username = var.username
}

provider "random" {}

data "ametnes_project" "project" {
  name = var.project_name
}

data "ametnes_location" "location" {
  name = var.location_name
  code = var.location_code
}

resource "ametnes_network" "network" {
  name     = var.network_name
  project  = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
}

data "ametnes_network" "network" {
  name       = var.network_name
  project    = data.ametnes_project.project.id
  location   = data.ametnes_location.location.id
  depends_on = [ametnes_network.network]
}

resource "random_string" "service_alias" {
  for_each = { for idx, svc in var.services : tostring(idx) => svc }
  length   = 5
  special  = false
  upper    = false
}

resource "ametnes_service" "service" {
  for_each    = { for idx, svc in var.services : tostring(idx) => svc }
  name        = "${each.value.kind_name}-service-${random_string.service_alias[each.key].result}"
  project     = data.ametnes_project.project.id
  location    = data.ametnes_location.location.id
  kind        = each.value.kind
  alias       = random_string.service_alias[each.key].result
  network     = data.ametnes_network.network.id
  capacity {
    storage = each.value.storage
  }
  config = {
    architecture = each.value.architecture
    "admin.email" = var.username
  }
  nodes = 1
}

output "service_connections" {
  value = { for k, svc in ametnes_service.service : k => svc.connections }
}
