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
  insecure = true
  username = "Brave.Microphone@ametnes.com"
}


resource "ametnes_network" "network" {
  name = "NETWORK-EUW5"
  project_name = "Demo"
  location = "d53059a1e6"
  description = "My loadbalance resource"
}

resource "ametnes_service" "grafana" {
  name = "grafana454"
  project_name = "Demo"
  location = "d53059a1e6"
  kind = "grafana:9.3"
  description = "sample grafana"
  network = 7540149073
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
