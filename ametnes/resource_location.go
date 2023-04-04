package ametnes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceLocation() *schema.Resource {
	return &schema.Resource{
		Description: `
Creates and manages an Ametnes Data Service location.
`,
		ReadContext:   dataSourceLocationsRead,
		CreateContext: resourceLocationCreate,
		DeleteContext: resourceLocationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "The name of the data service location.",

			},
			"code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "An easy identifier of the data service location. A combination of the `name` and `code` must be unique.",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "Description of the location.",
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "`true` if this location is enabled and can have resources created in it.",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The status of the data service location. This can be `OFFLINE` or `ONLINE`",
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
