package ametnes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
# curl -k -X GET "https://api-test.cloud.ametnes.com/v1/metadata/locations"
{
  "count": 2,
  "results": [
    {
      "create_date": "2020-12-19 09:02:15",
      "enabled": true,
      "id": "gcp.europe-west2",
      "name": "London, U.K.",
      "provider": "Google Cloud",
      "region": "gcp/europe-west2",
      "update_date": "2020-12-19 09:02:15"
    },
    {
      "create_date": "2020-12-19 09:02:15",
      "enabled": true,
      "id": "aws.eu-west-2",
      "name": "Europe (London)",
      "provider": "Amazon Web Service",
      "region": "aws/eu-west-2",
      "update_date": "2020-12-19 09:02:15"
    }
  ]
}
*/

func dataSourceLocation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceLocationsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"code": {
				Type:     schema.TypeString,
				Required: true,
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

func dataSourceLocationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	locationName := d.Get("name").(string)
	locationCode := d.Get("code").(string)

	locations, err := client.GetLocations()
	if err != nil {
		return diag.FromErr(err)
	}
	var foundLocation *Location
	for _, location := range locations {
		if location.Name == locationName && location.Location == locationCode {
			foundLocation = &location
			break
		}
	}

	if foundLocation == nil {
		return diag.Errorf("The location with code %s and name %s wasn't found", locationCode, locationName)
	}

	// always run
	d.SetId(foundLocation.Id)
	d.Set("location", foundLocation.Location)
	d.Set("enabled", foundLocation.Enabled)
	d.Set("status", foundLocation.Status)
	return nil
}
