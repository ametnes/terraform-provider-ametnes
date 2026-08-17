# Look up an existing location.
data "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

# Look up an existing project.
data "ametnes_project" "project" {
  name = "Demo"
}

# Provision multiple services using a map for stable resource addressing.
# The network attribute is optional — if omitted, a network is auto-created.
locals {
  services = {
    sentry = { kind = "sentry:26.2", kind_name = "sentry", storage = 30, architecture = "Starter" },
    matomo = { kind = "matomo:5.9",  kind_name = "matomo", storage = 10, architecture = "Starter" },
  }
}

resource "random_string" "service_alias" {
  for_each = local.services
  length   = 5
  special  = false
  upper    = false
}

resource "ametnes_service" "service" {
  for_each = local.services
  name     = "${each.value.kind_name}-service-${random_string.service_alias[each.key].result}"
  project  = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  kind     = each.value.kind
  alias    = random_string.service_alias[each.key].result
  capacity {
    storage = each.value.storage
  }
  config = {
    architecture  = each.value.architecture
    "admin.email" = var.username
  }
  nodes = 1
}

output "service_connections" {
  value = { for k, svc in ametnes_service.service : k => svc.connections }
}