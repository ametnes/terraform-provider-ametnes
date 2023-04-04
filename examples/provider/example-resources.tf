# Create a lcoation.
data "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

# Create project that will host all your resources.
resource "ametnes_project" "project" {
  name = "Demo"
}

# Create a network access resource.
resource "ametnes_network" "network" {
  name = "NETWORK-EUW5"
  project = ametnes_project.project.id
  location = data.ametnes_location.location.id
  description = "My loadbalance resource"
}

# Create a service resource.
resource "ametnes_service" "mysql" {
  name = "MySql-Demo-Instance"
  project = ametnes_project.project.id
  location = data.ametnes_location.location.id
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