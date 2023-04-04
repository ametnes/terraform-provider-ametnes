terraform {
  required_providers {
    ametnes = {
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

# Init and create the provider
provider "ametnes" {
  host = "https://cloud.ametnes.com/api/c/v1"
  token = var.token
  username = var.username

}

# Create a lcoation.
resource "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

