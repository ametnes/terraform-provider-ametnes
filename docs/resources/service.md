---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ametnes_service Resource - terraform-provider-ametnes"
subcategory: ""
description: |-
  
---

# ametnes_service (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `kind` (String)
- `location` (String)
- `name` (String)
- `nodes` (Number)
- `project` (String)

### Optional

- `capacity` (Block List, Max: 1) (see [below for nested schema](#nestedblock--capacity))
- `config` (Map of String)
- `description` (String)
- `network` (String)

### Read-Only

- `account` (String)
- `connections` (List of Object) (see [below for nested schema](#nestedatt--connections))
- `id` (String) The ID of this resource.
- `resource_id` (String)
- `status` (String)

<a id="nestedblock--capacity"></a>
### Nested Schema for `capacity`

Optional:

- `cpu` (Number)
- `memory` (Number)
- `storage` (Number)


<a id="nestedatt--connections"></a>
### Nested Schema for `connections`

Read-Only:

- `host` (String)
- `name` (String)
- `port` (String)

