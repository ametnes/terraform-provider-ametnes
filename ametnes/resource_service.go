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
		CreateContext: resourceServiceCreate,
		ReadContext:   resourceServiceRead,
		DeleteContext: resourceServiceDelete,

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
				Required: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cpu": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"storage": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"nodes": {
				Type:     schema.TypeInt,
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

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

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

	resource := Resource{
		Project:     projectID,
		Kind:        d.Get("kind").(string),
		Location:    d.Get("location").(string),
		Name:        d.Get("name").(string),
		Description: description,
		Spec: Spec{
			Components: map[string]interface{}{
				"cpu":     d.Get("cpu").(int),
				"storage": d.Get("storage").(int),
				"memory":  d.Get("memory").(int),
			},
			Networks: []Networks{
				{
					Id: d.Get("network").(int),
				},
			},
			Nodes: d.Get("nodes").(int),
		},
	}

	if net, ok := d.GetOk("network"); ok {
		resource.Network = net.(int)
	}
	service, err := client.CreateResource(resource)

	if err != nil {
		return diag.FromErr(err)
	}

	respChan := client.checkStatus(projectID, service.Id)
	select {
	case res := <-respChan:
		if res.Success {
			// Identity function
			d.SetId(fmt.Sprintf("%d/%d", projectID, service.Id))
			return resourceServiceRead(ctx, d, m)
		}
	case <-time.After(15 * time.Minute):
		return diag.Errorf("Timeout occured while checking for state")
	}

	// we will not reach here
	return nil
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

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
	d.Set("network", resource.Network)
	d.Set("status", resource.Status)
	d.Set("account", resource.Account)

	return nil
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	case <-time.After(10 * time.Minute):
		return diag.Errorf("Timeout occured while checking for state")
	}
	// we will not get here
	return nil
}
