terraform {
  required_providers {
    hashicups = {
      # version = "0.2"
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "hashicups" {}

module "psl" {
  source = "./location"

  location_name = "Packer Spiced Latte"
}

output "psl" {
  value = module.psl.locations
}
