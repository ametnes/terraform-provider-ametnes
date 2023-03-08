package ametnes

import (
	"context"
	"fmt"
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
		CreateContext: resourceNetworkCreate,
		ReadContext:   resourceServiceOrNetworkRead,
		DeleteContext: resourceServiceOrNetworkDelete,

		Schema: map[string]*schema.Schema{

			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // if the project name changes then we need to force new resource
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // if the name changes the we need to create a new resource
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"kind": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "loadbalancer:1.0",
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// computed
			"network": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
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

	projects, err := client.GetProjects()
	if err != nil {
		return diag.FromErr(err)
	}

	projectID := -1
	projectName := d.Get("project_name").(string)

	for _, project := range projects {
		if project.Name == projectName {
			projectID = project.Id
			break
		}
	}
	if projectID == -1 {
		return diag.Errorf("Cannot find your project with name %s", projectName)
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
