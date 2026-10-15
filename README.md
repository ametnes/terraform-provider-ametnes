# Terraform Provider Ametnes Cloud

The Ametnes Cloud provider lets you manage [Ametnes Cloud](https://cloud.ametnes.com/)
resources with Terraform: data service locations, projects, networks, and data services.

An Ametnes Data Service location is a dedicated Kubernetes cluster with an Ametnes Cloud
Agent installed. Data services are created and managed in the cluster using this provider.

## Requirements

- [Terraform](https://www.terraform.io/downloads) 0.13+
- [Go](https://golang.org/dl/) 1.18+ (to build the provider from source)

## Building and installing the provider

Build the provider binary:

```shell
go build -o terraform-provider-ametnes
```

Install it into your local Terraform plugins directory:

```shell
make install
```

## Authentication

To use this provider, generate an authentication token (aka API key) in your Ametnes Cloud
account: `User` -> `Edit` your user -> `Get User Token`. Keep the token secure - it will not
be visible again.

Two variables are required:

- `username` - the email address associated with your Ametnes Cloud account.
- `token` - the API token generated in the Ametnes console under your user account.

Supply them through any Terraform variable mechanism, for example `terraform.tfvars`:

```hcl
# terraform.tfvars
token    = "your-api-token"
username = "you@example.com"
```

## Example usage

### Configure the provider

```terraform
terraform {
  required_providers {
    ametnes = {
      source = "ametnes.com/cloud/ametnes"
    }
  }
}

provider "ametnes" {
  token    = var.token
  username = var.username
}

variable "token" {
  type = string
}

variable "username" {
  type = string
}
```

### Create a data service location

A data service location hosts your data services. Alternatively, create one in the
[Ametnes Cloud](https://cloud.ametnes.com/) console (`Service Locations` -> `New`) and note
the location id.

```terraform
resource "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"

  # Load balancer policy for services created in this location: only private
  # load balancers are allowed and new ones are private and cluster wide.
  network {
    public     = false
    private    = true
    visibility = "private"
    scope      = "location"
  }

  # Opt in to Premium/Enterprise telemetry. Omit the block to leave it disabled.
  telemetry {
    resources {
      logs    = true
      metrics = true
    }
    system {
      logs    = true
      metrics = true
    }
  }
}
```

> **On-premises:** load balancer visibility (`public`/`private`) is irrelevant because every
> load balancer is provisioned inside your own environment. Only the `scope` default applies.

### Create a project

All Ametnes Cloud resources must be created in a project.

```terraform
resource "ametnes_project" "project" {
  name        = "DemoProject"
  description = "Demo project"
}
```

### Create a network (load balancer)

A network access resource is the load balancer that exposes your data services. Load
balancers are **private by default**.

- `visibility` - `private` (default) or `public`.
- `scope` - `location` (default, cluster wide, shared by every project in the location) or
  `project` (only the owning project can attach).

When you omit `visibility` or `scope`, the location's defaults apply.

> **On-premises:** visibility is irrelevant because every load balancer is provisioned
> inside your own environment. Only `scope` applies.

```terraform
resource "ametnes_network" "network" {
  name        = "NETWORK-EUW1"
  project     = ametnes_project.project.id
  location    = ametnes_location.location.id
  description = "Cluster-wide private load balancer"
  visibility  = "private"
  scope       = "location"
}
```

### Create data services

Provision multiple services using a map for stable resource addressing. The `network`
attribute is optional: **when omitted, the Ametnes control plane manages the load balancer
for you**. It distributes services across the available load balancers, creating a new one
when needed and deleting ones that are no longer used, which keeps load balancer costs
optimal in public clouds (AWS, GKE, Azure).

The control plane allocates a service's standard port first (PostgreSQL `5432`, for
example) and falls back to the next free port when it is already taken - a second
PostgreSQL service on the same load balancer binds `5433`. Always read the port from the
service `connections` output rather than assuming the standard one.

Compute sizing (CPU/memory) is driven by the `architecture` preset in `config`. `storage`
is optional too - if omitted, the backend assigns a default value - and the value you set is
distributed across all the components that make up the service in predetermined proportions.

```terraform
# Look up an existing location.
data "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

# Look up an existing project.
data "ametnes_project" "project" {
  name = "Demo"
}

# Provision multiple services using a map for stable resource addressing.
# The network attribute is omitted - a load balancer is managed automatically.
locals {
  services = {
    sentry = { kind = "sentry:26.2", kind_name = "sentry", storage = 30, architecture = "Starter" },
    matomo = { kind = "matomo:5.9",  kind_name = "matomo", storage = 10, architecture = "Starter" },
  }
}

resource "random_string" "service_alias" {
  for_each = local.services
  length   = 5
  special  = false
  upper    = false
}

resource "ametnes_service" "service" {
  for_each = local.services
  name     = "${each.value.kind_name}-service-${random_string.service_alias[each.key].result}"
  project  = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
  kind     = each.value.kind
  alias    = random_string.service_alias[each.key].result
  capacity {
    storage = each.value.storage
  }
  config = {
    architecture  = each.value.architecture
    "admin.email" = var.username
  }
}

output "service_connections" {
  value = { for k, svc in ametnes_service.service : k => svc.connections }
}
```

### Attach a service to a specific load balancer

To pin a service to a load balancer you created, set the `network` attribute:

```terraform
resource "ametnes_service" "grafana" {
  name     = "grafana"
  project  = ametnes_project.project.id
  location = ametnes_location.location.id
  kind     = "grafana:9.3"
  network  = ametnes_network.network.id
  capacity {
    storage = 1
  }
}
```

### Look up existing resources

```terraform
data "ametnes_project" "project" {
  name = "Demo"
}

data "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

data "ametnes_network" "network" {
  name     = "NETWORK-EUW7"
  project  = data.ametnes_project.project.id
  location = data.ametnes_location.location.id
}

data "ametnes_kinds" "kinds" {
}
```

## Timeouts

The `timeouts` block controls how long the provider waits for a service or network to reach a
stable state. The actual deployment is performed asynchronously by the Ametnes Cloud Agent,
so the Terraform run must wait for the agent to finish before it completes.

Defaults for `ametnes_service`: `create` 60m, `update` 45m, `delete` 10m.

```terraform
timeouts {
  create = "60m"
  update = "2h"
  delete = "20m"
}
```

## Documentation

Full resource and data source reference:

- [Provider](docs/index.md)
- [ametnes_service](docs/resources/service.md)
- [ametnes_network](docs/resources/network.md)
- [ametnes_project](docs/resources/project.md)
- [ametnes_location](docs/resources/location.md)

## Development

Run the unit tests:

```shell
make testunit
```

Run the acceptance tests (requires `TF_ACC` and live credentials):

```shell
make testacc
```
