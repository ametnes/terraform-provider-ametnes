# Look up an existing location.
data "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

# Look up an existing project.
data "ametnes_project" "project" {
  name = "Demo"
}

# Provision multiple services from a list with random aliases.
# The network attribute is optional — if omitted, a network is auto-created.
locals {
  services = [
    { kind = "sentry:26.2", kind_name = "sentry", storage = 30, architecture = "Starter" },
    { kind = "matomo:5.9",  kind_name = "matomo", storage = 10, architecture = "Starter" },
  ]
}

resource "random_string" "service_alias" {
  for_each = { for idx, svc in local.services : tostring(idx) => svc }
  length   = 5
  special  = false
  upper    = false
}

resource "ametnes_service" "service" {
  for_each = { for idx, svc in local.services : tostring(idx) => svc }
  name     = "${each.value.kind_name}-service-${random_string.service_alias[each.key].result}"
  project  = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  kind     = each.value.kind
  alias    = random_string.service_alias[each.key].result
  capacity {
    storage = each.value.storage
    memory  = 1
    cpu     = 1
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