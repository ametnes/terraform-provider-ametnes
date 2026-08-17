variable "token" {
  type = string
}

variable "username" {
  type = string
}

variable "location_id" {
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