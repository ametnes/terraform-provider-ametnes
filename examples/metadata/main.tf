terraform {
  required_providers {
    ametnes = {
      # version = "0.3"
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

variable "location_name" {
  type    = string
  default = "Vagrante espresso"
}

data "ametnes_locations" "all" {}
data "ametnes_kinds" "all" {}

# Returns all locations
output "locations" {
  value = data.ametnes_locations.all.locations
}

output "kinds" {
  value = data.ametnes_kinds.all.kinds
}

# Only returns packer spiced latte
output "location" {
  value = {
    for coffee in data.ametnes_locations.all.locations :
    coffee.id => coffee
    if coffee.name == var.location_name
  }
}
