package ametnes

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"ametnes_service": resourceService(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ametnes_locations": dataSourceLocations(),
			"ametnes_kinds":     dataSourceKinds(),
		},
	}
}
