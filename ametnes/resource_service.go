package ametnes

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
			},
			"kind": {
				Type:     schema.TypeString,
				Required: true,
			},
			"product_code": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  DefaultProductCode,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},

			"cpu": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"storage": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"nodes": {
				Type:     schema.TypeInt,
				Required: true,
			},

			// computed
			"network": {
				Type:     schema.TypeInt,
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
	service, err := client.CreateResource(Resource{
		Project:     projectID,
		Kind:        d.Get("kind").(string),
		Location:    d.Get("location").(string),
		Name:        d.Get("name").(string),
		Product:     d.Get("product_code").(int),
		Description: description,
		Spec: Spec{
			Components: map[string]interface{}{
				"cpu":     d.Get("cpu").(int),
				"storage": d.Get("storage").(int),
				"memory":  d.Get("memory").(int),
			},
			Nodes: d.Get("nodes").(int),
		},
	},
	)

	if err != nil {
		return diag.FromErr(err)
	}
	// Identity function
	d.SetId(fmt.Sprintf("%d/%d", projectID, service.Id))
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
	return diag.FromErr(client.DeleteResource(Resource{
		Project: projectID,
		Id:      resourceID,
	}))
}