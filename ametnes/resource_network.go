package ametnes

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const DefaultStorage = 1
const DefaultNodes = 1

// Load balancer visibility and scope values understood by the Ametnes control
// plane. Visibility defaults to private; scope defaults to location (cluster
// wide) and may be narrowed to a single project.
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"

	ScopeLocation = "location"
	ScopeProject  = "project"
)

func resourceNetwork() *schema.Resource {
	return &schema.Resource{
		Description: `
Creates and manages a network access resource (a load balancer). Depending on your kubernetes cluster, this resource may be a load balancer or an object that manages a set of NodePort(s).

Load balancers are **private by default**. Use ` + "`visibility`" + ` to provision a public (` + "`public`" + `) or private (` + "`private`" + `) load balancer, and ` + "`scope`" + ` to control who can attach to it:

- ` + "`location`" + ` (default) makes the load balancer cluster wide: every project in the location can attach services to it.
- ` + "`project`" + ` scopes the load balancer to the owning project (namespace) only.

Terraform only overrides ` + "`visibility`" + ` and ` + "`scope`" + ` when you set them; otherwise the location defaults apply (see ` + "`ametnes_location`" + `). The ` + "`config`" + ` map carries advanced options such as ` + "`services_per_lb`" + `.

**On-premises:** visibility is irrelevant because every load balancer is provisioned inside your own environment.
`,
		CreateContext: resourceNetworkCreate,
		ReadContext:   resourceServiceOrNetworkRead,
		UpdateContext: resourceNetworkUpdate,
		DeleteContext: resourceServiceOrNetworkDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{

			"project": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // if the project name changes then we need to force new resource
				Description: "The `project` id of the project to create your network access resource.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The unique name of your network access resource.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of your network access resource.",
			},
			"kind": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "loadbalancer:1.0",
				Description: "The `kind` of your network access resource.",
			},
			"location": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The location id of your ametnes data service location to creat this network access resource in.",
			},
			"config": {
				Type:        schema.TypeMap,
				Required:    false,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Advanced configuration for your network access resource, for example `services_per_lb` (how many services to place on this load balancer before another is created). Use the dedicated `visibility` and `scope` attributes to control visibility and scope.",
			},
			"visibility": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{VisibilityPublic, VisibilityPrivate}, false),
				Description:  "Visibility of the load balancer. `private` (the default) provisions an internal load balancer reachable only inside the location; `public` provisions an internet facing load balancer. Public is only available at locations whose capabilities allow it. When omitted, the load balancer location default applies and the resolved value is recorded here. On-premises: visibility is irrelevant, because every load balancer is provisioned inside your own environment.",
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{ScopeLocation, ScopeProject}, false),
				Description:  "Scope of the load balancer. `location` (the default) makes the load balancer cluster wide so every project in the location can attach services to it; `project` scopes it to the owning project (namespace) only. When omitted, the load balancer location default applies and the resolved value is recorded here.",
			},

			// computed
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	client := m.(*Client)

	projectID, err := strconv.Atoi(d.Get("project").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	description := ""

	if desc, ok := d.GetOk("description"); ok {
		description = desc.(string)
	}
	kind := d.Get("kind").(string)

	config := networkSpecConfig(d)

	// we add network as prefix for network resource as thats how
	// server differentiates from other resources like service.
	prefixedKind := fmt.Sprintf("network/%s", kind)
	resource := Resource{
		Project:     projectID,
		Kind:        prefixedKind,
		Location:    d.Get("location").(string),
		Name:        d.Get("name").(string),
		Description: description,
		Spec: Spec{
			Components: map[string]interface{}{
				"storage": DefaultStorage,
			},
			Nodes:  DefaultNodes,
			Config: config,
		},
	}

	network, err := client.CreateResource(resource)

	if err != nil {
		return diag.FromErr(err)
	}

	respChan := client.checkStatus(projectID, network.Id)
	select {
	case res := <-respChan:
		if res.Success {
			d.SetId(fmt.Sprintf("%d/%d", projectID, network.Id))
			return resourceServiceOrNetworkRead(ctx, d, m)
		}
		if res.Error != nil {
			return diag.FromErr(res.Error)
		}
	case <-time.After(d.Timeout(schema.TimeoutCreate)):
		return diag.Errorf("Timeout occurred while checking for state")
	}

	return diag.Errorf("Unknown error while checking for state")
}

func resourceNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Check if any updatable fields have changed
	hasChanges := d.HasChange("name") || d.HasChange("description") || d.HasChange("config") ||
		d.HasChange("visibility") || d.HasChange("scope")

	if hasChanges {
		client := m.(*Client)

		projectID, err := strconv.Atoi(d.Get("project").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		ids := strings.Split(d.Id(), "/")
		resourceID, err := strconv.Atoi(ids[1])
		if err != nil {
			return diag.FromErr(err)
		}

		// Get the existing resource to preserve other fields
		existingResource, err := client.GetResource(projectID, resourceID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to get existing resource: %w", err))
		}

		// Get updated fields
		name := d.Get("name").(string)
		description := ""
		if desc, ok := d.GetOk("description"); ok {
			description = desc.(string)
		}

		// Get the config, always including the resolved visibility and scope
		config := networkSpecConfig(d)

		// Update the resource with new values
		updateResource := Resource{
			Id:          existingResource.Id,
			Project:     existingResource.Project,
			Account:     existingResource.Account,
			Kind:        existingResource.Kind,
			Location:    existingResource.Location,
			Network:     existingResource.Network,
			Name:        name,
			Status:      existingResource.Status,
			Description: description,
			Product:     existingResource.Product,
			Spec: Spec{
				Components: existingResource.Spec.Components,
				Nodes:      existingResource.Spec.Nodes,
				Config:     config,
				Networks:   existingResource.Spec.Networks,
			},
		}

		_, err = client.UpdateResource(updateResource)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update resource: %w", err))
		}

		respChan := client.checkStatus(projectID, resourceID)
		select {
		case res := <-respChan:
			if !res.Success {
				if res.Error != nil {
					return diag.FromErr(res.Error)
				}
				return diag.Errorf("Unknown error while checking for state after update")
			}
		case <-time.After(d.Timeout(schema.TimeoutUpdate)):
			return diag.Errorf("Timeout occurred while checking for state after update")
		}
	}

	// Read the updated resource state
	return resourceServiceOrNetworkRead(ctx, d, m)
}

// networkSpecConfig builds the spec config sent to the control plane. The user
// supplied `config` map is copied verbatim and the `visibility`/`scope`
// attributes are layered on top only when set, so an unset attribute lets the
// location default apply.
func networkSpecConfig(d *schema.ResourceData) map[string]interface{} {
	config := make(map[string]interface{})
	if v, ok := d.GetOk("config"); ok {
		for k, val := range v.(map[string]interface{}) {
			config[k] = val
		}
	}
	if v, ok := d.GetOk("visibility"); ok {
		config["visibility"] = v.(string)
	}
	if v, ok := d.GetOk("scope"); ok {
		config["scope"] = v.(string)
	}
	return config
}
