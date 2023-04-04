terraform {
  required_providers {
    ametnes = {
      source  = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  // add you provider here
  host = "https://cloud.ametnes.com/api/c/v1"
  token = var.token
  username = var.username

}

resource "ametnes_project" "project" {
  name = "Demo"
}

resource "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

resource "ametnes_network" "network" {
  name = "NETWORK-EUW5"
  project = ametnes_project.project.id
  location = ametnes_location.location.id
  description = "My loadbalance resource"
}

resource "ametnes_service" "mysql" {
  name = "MySql-Demo-Instance"
  project = ametnes_project.project.id
  location = ametnes_location.location.id
  kind = "mysql:8.0"
  description = "Mysql Demo Instance"
  network = ametnes_network.network.resource_id
  capacity {
    storage = 10
    memory = 1
    cpu = 1
  } 
  nodes = 1
}


output "mysql_connections" {
  value = ametnes_service.mysql.connections
}
