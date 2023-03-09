terraform {
  required_providers {
    ametnes = {
      # version = "0.3"
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  // add you provider here
  host = "https://api-test.cloud.ametnes.com/v1"
  token = var.token
  insecure = true
  username = "Brave.Microphone@ametnes.com"
}

data "ametnes_project" "project" {
  name = "Demo"
}

data "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

resource "ametnes_network" "network" {
  name = "NETWORK-EUW5"
  project = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  description = "My loadbalance resource"
}
