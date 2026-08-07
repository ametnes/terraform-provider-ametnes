package ametnes

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetwork() *schema.Resource {
	return &schema.Resource{
		Description: `
Read a network access resource.
`,
		ReadContext: dataSourceNetworksRead,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Ametnes cloud project of the network access resource.",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Name of the network access resource.",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Description of the network access resource.",
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Location of the network access resource.",
			},
			"kind": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Kind of the network access resource.",
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
		// Check if this is an authentication error
		errMsg := err.Error()
		if strings.Contains(errMsg, "401") || 
			strings.Contains(errMsg, "Unauthorized") ||
			strings.Contains(errMsg, "status: 401") {
			return diag.Errorf("Authentication failed: %v", err)
		}
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

	d.SetId(fmt.Sprintf("%d/%d", projectID, foundResource.Id))
	d.Set("kind", foundResource.Kind)
	return nil
}
