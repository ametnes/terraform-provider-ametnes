variable "token" {
  type = string
}

variable "username" {
  type = string
}

variable "hdb_admin_user" {
  type = string
  sensitive = true
}

variable "hdb_admin_pass" {
  type = string
  sensitive = true
}

variable "hdb_clustering_user" {
  type = string
  sensitive = true
}

variable "hdb_clustering_pass" {
  type = string
  sensitive = true
}
