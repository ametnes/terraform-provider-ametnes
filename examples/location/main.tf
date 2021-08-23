terraform {
  required_providers {
    ametnes = {
      # version = "0.2"
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

variable "location_name" {
  type    = string
  default = "Vagrante espresso"
}

data "ametnes_locations" "all" {}

# Returns all locations
output "locations" {
  value = data.ametnes_locations.all.locations
}

# Only returns packer spiced latte
output "location" {
  value = {
    for coffee in data.ametnes_locations.all.locations :
    coffee.id => coffee
    if coffee.name == var.location_name
  }
}
