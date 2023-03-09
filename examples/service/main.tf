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

data "ametnes_network" "network" {
  name = "NETWORK-EUW5"
  project = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
}

resource "ametnes_service" "grafana" {
  name = "grafana455"
  project = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  kind = "grafana:9.3"
  description = "sample grafana"
  network = data.ametnes_network.network.id
  capacity {
    storage = 1
    memory = 1
    cpu = 1
  }

  config = {
    "auth.azuread.client_id" = "SomeText"
    "auth.azuread.client_secret" = "SomeText"
  }
 
  nodes = 1
}
