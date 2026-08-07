terraform {
  required_providers {
    ametnes = {
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

# Init and create the provider
provider "ametnes" {
  token = var.token
  username = var.username

}

# Create a location.
resource "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

