terraform {
  required_providers {
    ametnes = {
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  token = var.token
  username = var.username
}

resource "ametnes_location" "location" {
  name = "Acme"
  code = "USE1"
  description = "US-East-1 Data Service Location"
}
