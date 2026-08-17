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

const DefaultProductCode = 3795211474

func resourceService() *schema.Resource {
	return &schema.Resource{
		Description: `
Creates and manages a data service resource. All data service resources created must be attached to a network resource.
`,
		CreateContext: resourceServiceCreate,
		ReadContext:   resourceServiceOrNetworkRead,
		DeleteContext: resourceServiceOrNetworkDelete,
		UpdateContext: resourceServiceUpdate,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(45 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{

			"project": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // if the project changes then we need to force new resource
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
				Required:    true,
				ForceNew:    true,
				Description: "The `kind` of your data service resource. Examples: `grafana:8.3`, `harperdb:3.3`",
			},
			"location": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The location id of your ametnes data service location to creat this data service resource in.",
			},
			"capacity": {
				Type:        schema.TypeList,
				Required:    false,
				Optional:    true,
				MaxItems:    1,
				Description: "Capacity specs for your data service resource. Only `storage` is configurable.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Storage, in (Gb) unit counts, for your data service resource. Optional — if omitted, the backend assigns a default value. The value is distributed across all components that make up the service in predetermined proportions.",
						},
					},
				},
			},
			"config": {
				Type:        schema.TypeMap,
				Required:    false,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Configuration details for your data service resource.",
			},
			"alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An optional alias for your data service resource.",
			},
			// computed
			"network": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Network resource your data service resource will be attached to in order to expose it.",
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
			"connections": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Connection details to access your data service resource.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

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

	// get the capacity
	capacity, err := expandCapacitySchema(d.Get("capacity").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	var config map[string]interface{}
	if v, ok := d.GetOk("config"); ok {
		config = v.(map[string]interface{})
	}
	if config == nil {
		config = make(map[string]interface{})
	}
	if _, ok := config["public.visible"]; !ok {
		config["public.visible"] = "true"
	}
	d.Set("config", config)

	alias := ""
	if a, ok := d.GetOk("alias"); ok {
		alias = a.(string)
	}

	// we add service as prefix for service resource as thats how
	// server differentiates from other resources like network.
	prefixedKind := fmt.Sprintf("service/%s", kind)
	resource := Resource{
		Project:     projectID,
		Kind:        prefixedKind,
		Location:    d.Get("location").(string),
		Name:        d.Get("name").(string),
		Description: description,
		Alias:       alias,
		Spec: Spec{
			Components: map[string]interface{}{
				"storage": capacity.Storage,
			},
			Nodes:  DefaultNodes,
			Config: config,
		},
	}

	if networkIntf, ok := d.GetOk("network"); ok {
		networkStr := networkIntf.(string)

		// if there is a project id present as a prefix lets
		// just remove it
		if strings.Contains(networkStr, "/") {
			networkParts := strings.Split(networkStr, "/")
			// we remove the project id part
			networkStr = networkParts[1]
		}
		networkInt, err := strconv.Atoi(networkStr)
		if err != nil {
			return diag.FromErr(err)
		}
		resource.Network = networkInt
		resource.Spec.Networks = []Networks{
			{
				Id: networkInt,
			},
		}
	}
	service, err := client.CreateResource(resource)

	if err != nil {
		return diag.FromErr(err)
	}

	respChan := client.checkStatus(projectID, service.Id)
	select {
	case res := <-respChan:
		if res.Success {
			d.SetId(fmt.Sprintf("%d/%d", projectID, service.Id))
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

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Prevent changing the network field
	if d.HasChange("network") {
		return diag.Errorf("Changing the 'network' field is not permitted for this resource. You must destroy and recreate the resource to use a different network.")
	}

	// Check if any updatable fields have changed
	hasChanges := d.HasChange("name") || d.HasChange("description") || d.HasChange("capacity") ||
		d.HasChange("config") || d.HasChange("alias")

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

		alias := ""
		if a, ok := d.GetOk("alias"); ok {
			alias = a.(string)
		}

		// Get the capacity
		capacity, err := expandCapacitySchema(d.Get("capacity").([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		// Get the config
		var config map[string]interface{}
		if v, ok := d.GetOk("config"); ok {
			config = v.(map[string]interface{})
		}
		if config == nil {
			config = make(map[string]interface{})
		}
		if _, ok := config["public.visible"]; !ok {
			config["public.visible"] = "true"
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
			Alias:       alias,
			Status:      existingResource.Status,
			Description: description,
			Product:     existingResource.Product,
			Spec: Spec{
				Components: map[string]interface{}{
					"storage": capacity.Storage,
				},
				Nodes:    DefaultNodes,
				Config:   config,
				Networks: existingResource.Spec.Networks,
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

func resourceServiceOrNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

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

	resource, err := client.GetResource(projectID, resourceID)
	if err != nil {
		// if we get error while getting resource then
		d.SetId("")
		return nil
	}
	d.Set("resource_id", fmt.Sprint(resource.Id))
	d.Set("status", resource.Status)
	d.Set("account", fmt.Sprint(resource.Account))
	d.Set("name", resource.Name)
	d.Set("description", resource.Description)
	d.Set("alias", resource.Alias)

	// Set network field only for service resources
	if strings.HasPrefix(resource.Kind, "service/") {
		d.Set("network", fmt.Sprintf("%d/%d", projectID, resource.Network))
	}

	// Set service-specific fields only for service resources
	if strings.HasPrefix(resource.Kind, "service/") {
		// Set capacity
		if resource.Spec.Components != nil {
			capacity := []map[string]interface{}{
				{
					"storage": resource.Spec.Components["storage"],
				},
			}
			d.Set("capacity", capacity)
		}
	}

	// Set config if it exists (for both service and network resources)
	if strings.HasPrefix(resource.Kind, "service/") || strings.HasPrefix(resource.Kind, "network/") {
		if resource.Spec.Config != nil && len(resource.Spec.Config) > 0 {
			// Convert map[string]interface{} to map[string]string for schema
			configMap := make(map[string]string)
			for k, v := range resource.Spec.Config {
				configMap[k] = fmt.Sprintf("%v", v)
			}
			d.Set("config", configMap)
		} else {
			// If config is empty or nil, set empty map
			d.Set("config", make(map[string]string))
		}
	}

	connections := []Connection{}
	if resource.Spec.Connections != nil && len(resource.Spec.Connections) != 0 {
		connections = resource.Spec.Connections
	} else if resource.Spec.Connection.Host != "" {
		connections = append(connections, resource.Spec.Connection)
	}

	var conns []interface{}
	for _, conn := range connections {
		connMap := make(map[string]interface{})
		connMap["name"] = conn.Name
		connMap["host"] = conn.Host
		connMap["port"] = fmt.Sprint(conn.Port)

		conns = append(conns, connMap)
	}

	if len(connections) != 0 {
		d.Set("connections", conns)
	}

	return nil
}

func resourceServiceOrNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	err = client.DeleteResource(Resource{
		Project: projectID,
		Id:      resourceID,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	respChan := client.checkStatusDelete(projectID, resourceID)
	select {
	case res := <-respChan:
		if res.Success {
			// Identity function
			d.SetId("")
			return nil
		}
	case <-time.After(d.Timeout(schema.TimeoutDelete)):
		return diag.Errorf("Timeout occurred while checking for state")
	}
	// we will not get here
	return nil
}

func expandCapacitySchema(in []interface{}) (*Capacity, error) {
	cap := &Capacity{}
	if len(in) == 0 || in[0] == nil {
		return &Capacity{Storage: 1}, nil
	}
	m := in[0].(map[string]interface{})

	if storage, ok := m["storage"]; ok {
		cap.Storage = storage.(int)
	} else {
		cap.Storage = 1
	}

	return cap, nil
}
