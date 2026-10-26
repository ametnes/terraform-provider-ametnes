terraform {
  required_providers {
    ametnes = {
      source = "ametnes.com/cloud/ametnes"
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

# Derive the location id from the location's unique name and code.
data "ametnes_location" "location" {
  name = var.location_name
  code = var.location_code
}

# The network attribute on the service is optional.
# If omitted, the Ametnes control plane manages a load balancer automatically.

resource "random_string" "service_alias" {
  for_each = var.services
  length   = 5
  special  = false
  upper    = false
}

resource "ametnes_service" "service" {
  for_each = var.services
  name     = "${each.value.kind_name}-service-${random_string.service_alias[each.key].result}"
  project  = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  kind     = each.value.kind
  alias    = random_string.service_alias[each.key].result
  capacity {
    storage = each.value.storage
  }
  config = {
    architecture     = each.value.architecture
    "admin.email"    = var.username
    "admin.user"     = var.admin_user
    "admin.password" = var.admin_password
    # SMTP relay for services that send email (e.g. langfuse, sentry, gitlab,
    # matomo). "ametnes" uses the Ametnes-managed relay. To use your own
    # provider, pass a connection string instead, e.g.
    # "smtp://apikey:KEY@smtp.sendgrid.net:587?from=name%40example.com"
    # (percent-encode the password) or "smtps://user:pass@smtp.example.com:465".
    # Remove the key to disable email.
    "smtp" = "ametnes"
  }
  timeouts {
    create = "3h"
    update = "3h"
    delete = "20m"
  }
}

output "service_connections" {
  value = { for k, svc in ametnes_service.service : k => svc.connections }
}
