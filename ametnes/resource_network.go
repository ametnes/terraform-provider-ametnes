package ametnes

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const DefaultCpu = 1
const DefaultStorage = 1
const DefaultMemory = 1
const DefaultNodes = 1

func resourceNetwork() *schema.Resource {
	return &schema.Resource{
		Description: `
Creates and manages a network access resource. Depending on your kubernetes cluster, this resource may be a load balancer or an object that manages a set of NodePort(s).
`,
		CreateContext: resourceNetworkCreate,
		ReadContext:   resourceServiceOrNetworkRead,
		UpdateContext: resourceNetworkUpdate,
		DeleteContext: resourceServiceOrNetworkDelete,

		Schema: map[string]*schema.Schema{

			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // if the project name changes then we need to force new resource
				Description: "The `project` id of the project to create your network access resource.",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: "The unique name of your network access resource.",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The description of your network access resource.",
			},
			"kind": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "loadbalancer:1.0",
				Description: "The `kind` of your network access resource.",
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "The location id of your ametnes data service location to creat this network access resource in.",
			},

			// computed
			"network": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
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
				"cpu":     DefaultCpu,
				"storage": DefaultStorage,
				"memory":  DefaultMemory,
			},
			Nodes: DefaultNodes,
		},
	}

	if net, ok := d.GetOk("network"); ok {
		resource.Network = net.(int)
	}
	network, err := client.CreateResource(resource)

	if err != nil {
		return diag.FromErr(err)
	}

	respChan := client.checkStatus(projectID, network.Id)
	select {
	case res := <-respChan:
		if res.Success {
			// Identity function
			d.SetId(fmt.Sprintf("%d/%d", projectID, network.Id))
			return resourceServiceOrNetworkRead(ctx, d, m)
		}
	case <-time.After(15 * time.Minute):
		return diag.Errorf("Timeout occured while checking for state")
	}

	// we will not reach here
	return nil
}

func resourceNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Check if any updatable fields have changed
	hasChanges := d.HasChange("name") || d.HasChange("description")

	if hasChanges {
		client := m.(*Client)

		ids := strings.Split(d.Id(), "/")
		projectID, err := strconv.Atoi(ids[0])
		if err != nil {
			return diag.FromErr(err)
		}
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
				Config:     existingResource.Spec.Config,
				Networks:   existingResource.Spec.Networks,
			},
		}

		_, err = client.UpdateResource(updateResource)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update resource: %w", err))
		}
	}

	// Read the updated resource state
	return resourceServiceOrNetworkRead(ctx, d, m)
}
