package ametnes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const AMETNES_HOST = "https://cloud.ametnes.com/api"

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AMETNES_TOKEN", nil),
				Description: "The API token for API operations.",
				Sensitive:   true,
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AMETNES_USERNAME", nil),
				Description: "The username for API operations.",
			},
		},
		ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			token := data.Get("token").(string)
			username := data.Get("username").(string)
			client, err := NewClient(AMETNES_HOST, Token{
				Type:     Basic,
				Token:    token,
				Username: &username,
			})

			if err != nil {
				return nil, diag.FromErr(err)
			}

			return client, nil
		},
		ResourcesMap: map[string]*schema.Resource{
			"ametnes_service": resourceService(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"ametnes_locations": dataSourceLocations(),
			"ametnes_kinds":     dataSourceKinds(),
		},
	}
}
