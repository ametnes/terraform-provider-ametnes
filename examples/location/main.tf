terraform {
  required_providers {
    ametnes = {
      source = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  token    = var.token
  username = var.username
}

# Create a data service location and configure its load balancer policy
# (capabilities and defaults) and the telemetry the agent collects.
resource "ametnes_location" "location" {
  name        = var.location_name
  code        = var.location_code
  description = "Example Data Service location"

  # capability - the load balancer visibilities services at this location are
  # allowed to request. default - what a load balancer uses when a service or
  # network does not choose one.
  #
  # On-premises: visibility (public/private) is irrelevant because every load
  # balancer is provisioned inside your own environment; only scope applies.
  network {
    public     = var.load_balancer_public
    private    = var.load_balancer_private
    visibility = var.default_visibility
    scope      = var.default_scope
  }

  # Telemetry the agent collects. Requires a Premium or Enterprise plan; omit
  # the block to leave telemetry disabled.
  telemetry {
    resources {
      logs    = var.telemetry_resources_logs
      metrics = var.telemetry_resources_metrics
    }
    system {
      logs    = var.telemetry_system_logs
      metrics = var.telemetry_system_metrics
    }
  }
}

output "location_id" {
  value = ametnes_location.location.id
}
