terraform {
  required_providers {
    ametnes = {
      # version = "0.2"
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {}

module "psl" {
  source = "./metadata"

  location_name = "Packer Spiced Latte"
}

output "psl" {
  value = module.psl.locations
}
output "ps2" {
  value = module.psl.kinds
}
