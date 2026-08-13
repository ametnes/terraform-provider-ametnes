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

# The network attribute on the service is optional.
# If omitted, a network resource is automatically created.

resource "random_string" "service_alias" {
  for_each = var.services
  length   = 5
  special  = false
  upper    = false
}

resource "ametnes_service" "service" {
  for_each    = var.services
  name        = "${each.value.kind_name}-service-${random_string.service_alias[each.key].result}"
  project     = data.ametnes_project.project.id
  location    = var.location_id
  kind        = each.value.kind
  alias       = random_string.service_alias[each.key].result
  capacity {
    storage = each.value.storage
  }
  config = {
    architecture = each.value.architecture
    "admin.email" = var.username
    "public.visible" = "true"
  }
  nodes = 1
  timeouts {
    create = "3h"
    update = "3h"
    delete = "20m"
  }
}

output "service_connections" {
  value = { for k, svc in ametnes_service.service : k => svc.connections }
}
