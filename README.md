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
account: `User` -> `Edit` your user -> `Get User Token`. Keep the token secure — it will not
be visible again.

Two variables are required:

- `username` — the email address associated with your Ametnes Cloud account.
- `token` — the API token generated in the Ametnes console under your user account.

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
      source  = "ametnes.com/cloud/ametnes"
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
}
```

### Create a project

All Ametnes Cloud resources must be created in a project.

```terraform
resource "ametnes_project" "project" {
  name        = "DemoProject"
  description = "Demo project"
}
```

### Create a network

A network access resource exposes your data services. Depending on your Kubernetes cluster
this may be a load balancer or a set of `NodePort`s. The `config` map accepts a `public` key:

- `"public" = "true"` — provisions a public-facing load balancer (default behavior).
- `"public" = "false"` — provisions a private load balancer accessible only within the network.

```terraform
resource "ametnes_network" "network" {
  name        = "NETWORK-EUW8"
  project     = data.ametnes_project.project.id
  location    = data.ametnes_location.location.id
  description = "My load balancer resource"
  config = {
    "public" = "true"
  }
}
```

### Create data services

Provision multiple services using a map for stable resource addressing. The `network`
attribute is optional — if omitted, a network is automatically created. Compute sizing
(CPU/memory) is driven by the `architecture` preset in `config`. `storage` is optional too
— if omitted, the backend assigns a default value — and the value you set is distributed
across all the components that make up the service in predetermined proportions.

```terraform
terraform {
  required_providers {
    ametnes = {
      source  = "ametnes.com/cloud/ametnes"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

# Look up an existing location.
data "ametnes_location" "location" {
  name = "Ametnes Cloud"
  code = "EUW1"
}

# Look up an existing project.
data "ametnes_project" "project" {
  name = "Demo"
}

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

### Look up existing resources

```terraform
data "ametnes_project" "project" {
  name = "Default"
}

data "ametnes_location" "location" {
  name = "Ametnes"
  code = "DSL-USE1"
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
