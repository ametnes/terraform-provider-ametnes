package ametnes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLocation() *schema.Resource {
	return &schema.Resource{
		ReadContext:   dataSourceLocationsRead,
		CreateContext: resourceLocationCreate,
		DeleteContext: resourceLocationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceLocationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	locationName := d.Get("name").(string)
	locationCode := d.Get("code").(string)
	locationDescription := ""

	if description, ok := d.GetOk("description"); ok {
		locationDescription = description.(string)
	}

	createdLocation, err := client.CreateLocation(Location{
		Name:        locationName,
		Location:    locationCode,
		Description: locationDescription,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(createdLocation.Id)

	return dataSourceLocationsRead(ctx, d, m)

}

func resourceLocationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	locationID := d.Id()

	err := client.DeleteLocation(Location{
		Id: locationID,
	})

	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
