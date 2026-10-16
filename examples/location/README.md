# Example: ametnes_location

Creates an Ametnes Data Service location and configures its **load balancer
policy** (capabilities and defaults) and the **telemetry** the location's agent
collects.

## Usage

```shell
terraform init
terraform plan
terraform apply
```

`terraform.tfvars` holds the credentials and settings. Replace `username` and
`token` with your Ametnes Cloud account values.

## Load balancer policy (`network` block)

Capabilities are the visibilities services at the location are *allowed* to
request; defaults are what a load balancer uses when a service or network does
*not* choose one.

| Setting      | Meaning                                                              | Default    |
| ------------ | -------------------------------------------------------------------- | ---------- |
| `public`     | Allow public (internet facing) load balancers                        | `false`    |
| `private`    | Allow private (internal) load balancers                              | `true`     |
| `visibility` | Default visibility (`private` or `public`)                           | `private`  |
| `scope`      | Default scope (`location` cluster wide, or `project` namespace scoped) | `location` |

On-premises: visibility is irrelevant because every load balancer is provisioned
inside your own environment; only `scope` applies.

## Telemetry (`telemetry` block)

`resources` and `system` scopes each collect `logs` and `metrics`. Requires a
Premium or Enterprise plan; omit the block to leave telemetry disabled.
