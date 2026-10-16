variable "token" {
  type = string
}

variable "username" {
  type = string
}

variable "admin_user" {
  type = string
}

variable "admin_password" {
  type      = string
  sensitive = true
}

variable "location_code" {
  type = string
}

variable "location_name" {
  type = string
}

variable "project_name" {
  type = string
}

variable "services" {
  type = map(object({
    kind         = string
    kind_name    = string
    storage      = number
    architecture = string
  }))
}