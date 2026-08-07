variable "token" {
  type = string
}

variable "username" {
  type = string
}

variable "location_name" {
  type = string
}

variable "location_code" {
  type = string
}

variable "network_name" {
  type = string
}

variable "project_name" {
  type = string
}

variable "services" {
  type = list(object({
    kind         = string
    kind_name    = string
    storage      = number
    architecture = string
  }))
}