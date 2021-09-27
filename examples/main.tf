terraform {
  required_providers {
    ametnes = {
      # version = "0.2"
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  // add you provider here
}

resource "ametnes_service" "sample" {
  name = "Test Sample"
  project_name = "Default"
  location = "gcp/europe-west2"
  kind = "service/mysql:8.0"
  description = "My name is Sample"
  storage = 1
  memory = 1
  nodes = 1
  cpu = 1
}
