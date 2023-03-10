package ametnes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworksRead,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"kind": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceNetworksRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	projectstr := d.Get("project").(string)
	location := d.Get("location").(string)
	name := d.Get("name").(string)
	projectID, _ := strconv.Atoi(projectstr)

	resources, err := client.GetResources(&Project{
		Id: projectID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var foundResource *Resource
	for _, resource := range resources {
		if resource.Location == location && resource.Name == name {
			foundResource = &resource
			break
		}
	}

	if foundResource == nil {
		return diag.Errorf("Cannot found a network resource based on information provided")
	}

	d.SetId(fmt.Sprint(foundResource.Id))
	d.Set("kind", foundResource.Kind)
	return nil
}
