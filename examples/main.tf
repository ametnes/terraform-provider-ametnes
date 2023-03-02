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
  host = "https://api-test.cloud.ametnes.com/v1"
  insecure = true
}

resource "ametnes_service" "network" {
  name = "NETWORK-EUW2"
  project_name = "Default"
  location = "aws/eu-west-2"
  kind = "network/privatelink:1.0"
  description = "My name is Sample"
  storage = 1
  memory = 1
  nodes = 1
  cpu = 1
}

resource "ametnes_service" "mysql" {
  count = 3
  name = "Test Sample ${count.index}"
  project_name = "Default"
  location = "aws/eu-west-2"
  kind = "service/mysql:8.0"
  description = "My name is Sample"
  network = tonumber(element(split("/", ametnes_service.network.id), 1))
  storage = 1
  memory = 1
  nodes = 1
  cpu = 1
}

resource "ametnes_service" "postgres" {
  count = 4
  name = "MySql Db ${count.index}"
  project_name = "Default"
  location = "aws/eu-west-2"
  kind = "service/mysql:8.0"
  description = "My name is Sample"
  network = tonumber(element(split("/", ametnes_service.network.id), 1))
  storage = 1
  memory = 1
  nodes = 1
  cpu = 1
}

resource "ametnes_service" "pg" {
  count = 4
  name = "Postgres DB ${count.index}"
  project_name = "Default"
  location = "gcp/europe-west2"
  kind = "service/postgres:11.9"
  description = "My name is Sample"
  storage = 1
  memory = 1
  nodes = 1
  cpu = 1
}