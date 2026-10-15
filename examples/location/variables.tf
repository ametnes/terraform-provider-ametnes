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

# Load balancer capability: allow public (internet facing) load balancers.
variable "load_balancer_public" {
  type    = bool
  default = false
}

# Load balancer capability: allow private (internal) load balancers.
variable "load_balancer_private" {
  type    = bool
  default = true
}

# Default load balancer visibility used when a service or network does not
# choose one (`public` or `private`).
variable "default_visibility" {
  type    = string
  default = "private"
}

# Default load balancer scope used when a service or network does not choose one
# (`location` = cluster wide, or `project` = project/namespace scoped).
variable "default_scope" {
  type    = string
  default = "location"
}

# Telemetry: logs and metrics for the data services (resources). Premium or
# Enterprise plan required.
variable "telemetry_resources_logs" {
  type    = bool
  default = false
}

variable "telemetry_resources_metrics" {
  type    = bool
  default = false
}

# Telemetry: logs and metrics for the location's system components. Premium or
# Enterprise plan required.
variable "telemetry_system_logs" {
  type    = bool
  default = false
}

variable "telemetry_system_metrics" {
  type    = bool
  default = false
}
