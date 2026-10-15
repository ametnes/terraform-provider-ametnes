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

data "ametnes_project" "project" {
  name = "Default"
}

data "ametnes_location" "location" {
  name = "Ametnes"
  code = "DSL-USE1"
}

# A private, cluster-wide load balancer (the defaults). Services from any
# project in the location can attach to it.
#
# On-premises: visibility is irrelevant because every load balancer is
# provisioned inside your own environment.
resource "ametnes_network" "network" {
  name        = "NETWORK-EUW8"
  project     = data.ametnes_project.project.id
  location    = data.ametnes_location.location.id
  description = "My load balancer resource"
  visibility  = "private"
  scope       = "location"
}

# A public, project-scoped load balancer. Only services in this project can
# attach to it. Public requires a location that allows public load balancers.
resource "ametnes_network" "public_project_network" {
  name        = "NETWORK-PUBLIC-EUW8"
  project     = data.ametnes_project.project.id
  location    = data.ametnes_location.location.id
  description = "Public, project scoped load balancer"
  visibility  = "public"
  scope       = "project"
  config = {
    "services_per_lb" = "3"
  }
}
