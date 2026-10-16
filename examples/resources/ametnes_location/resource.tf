terraform {
  required_providers {
    ametnes = {
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  token    = var.token
  username = var.username
}

# A private-only location that collects resource and system telemetry.
resource "ametnes_location" "location" {
  name        = "Acme"
  code        = "USE1"
  description = "US-East-1 Data Service Location"

  # Load balancer policy: only private load balancers are allowed, and new load
  # balancers are private and cluster wide unless a service chooses otherwise.
  # On-premises: visibility is irrelevant because every load balancer is
  # provisioned inside your own environment; only scope applies.
  network {
    public     = false
    private    = true
    visibility = "private"
    scope      = "location"
  }

  # Premium/Enterprise telemetry. Omit the block to leave telemetry disabled.
  telemetry {
    resources {
      logs    = true
      metrics = true
    }
    system {
      logs    = true
      metrics = true
    }
  }
}
