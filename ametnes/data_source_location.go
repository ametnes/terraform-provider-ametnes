package ametnes

import (
	"context"
	"strings"

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
		Description: `
Read a data service location
`,
		ReadContext: dataSourceLocationsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the data source.",
			},
			"code": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique code of the data source.",
			},
			"location": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Alternative ocation unique code.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "`true` if the location is enabled.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the location such as `OFFLINE`, `ONLINE`.",
			},
			"network": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Load balancer policy for the location: allowed visibilities (`public`, `private`) and defaults (`visibility`, `scope`).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"public": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the location allows public (internet facing) load balancers.",
						},
						"private": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the location allows private (internal) load balancers.",
						},
						"visibility": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Default load balancer visibility at the location (`public` or `private`).",
						},
						"scope": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Default load balancer scope at the location (`location` or `project`).",
						},
					},
				},
			},
			"telemetry": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Telemetry the location's agent collects.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": telemetryScopeComputedSchema(),
						"system":    telemetryScopeComputedSchema(),
					},
				},
			},
		},
	}
}

// telemetryScopeComputedSchema is the read-only schema for the `resources` and
// `system` telemetry scopes exposed by the location data source.
func telemetryScopeComputedSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"logs": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Whether logs are collected.",
				},
				"metrics": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Whether metrics are collected.",
				},
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
		// Check if this is an authentication error
		errMsg := err.Error()
		if strings.Contains(errMsg, "401") ||
			strings.Contains(errMsg, "Unauthorized") ||
			strings.Contains(errMsg, "status: 401") {
			return diag.Errorf("Authentication failed: %v", err)
		}
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
	setLocationAttributes(d, foundLocation.Attributes)
	return nil
}

// setLocationAttributes copies the load balancer policy and telemetry values the
// cloud API reports into the schema shared by the location resource and data
// source.
func setLocationAttributes(d *schema.ResourceData, attributes map[string]interface{}) {
	if attributes == nil {
		return
	}

	if _, ok := locationAttribute(attributes, "capabilities"); ok {
		d.Set("network", flattenNetwork(attributes))
	}

	if _, ok := locationAttribute(attributes, "telemetry"); ok {
		d.Set("telemetry", flattenTelemetry(attributes))
	}
}

// flattenNetwork builds the `network` block from the API attributes, treating a
// missing capability or default as the platform default.
func flattenNetwork(attributes map[string]interface{}) []interface{} {
	public := false
	if v, ok := locationAttribute(attributes, "capabilities", "network", "load_balancer", "public"); ok {
		public, _ = v.(bool)
	}
	private := true
	if v, ok := locationAttribute(attributes, "capabilities", "network", "load_balancer", "private"); ok {
		private, _ = v.(bool)
	}
	visibility := DefaultLoadBalancerVisibility
	if v, ok := locationAttribute(attributes, "defaults", "network", "load_balancer", "visibility"); ok {
		if s, ok := v.(string); ok && s != "" {
			visibility = s
		}
	}
	scope := DefaultLoadBalancerScope
	if v, ok := locationAttribute(attributes, "defaults", "network", "load_balancer", "scope"); ok {
		if s, ok := v.(string); ok && s != "" {
			scope = s
		}
	}
	return []interface{}{
		map[string]interface{}{
			"public":     public,
			"private":    private,
			"visibility": visibility,
			"scope":      scope,
		},
	}
}

// flattenTelemetry builds the telemetry block from the API attributes, treating
// a missing scope or feature as disabled.
func flattenTelemetry(attributes map[string]interface{}) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"resources": flattenTelemetryScope(attributes, "resources"),
			"system":    flattenTelemetryScope(attributes, "system"),
		},
	}
}

func flattenTelemetryScope(attributes map[string]interface{}, scope string) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"logs":    locationBool(attributes, "telemetry", scope, "logs", "enabled"),
			"metrics": locationBool(attributes, "telemetry", scope, "metrics", "enabled"),
		},
	}
}

// locationBool reads a nested boolean, returning false when it is absent.
func locationBool(attributes map[string]interface{}, path ...string) bool {
	value, ok := locationAttribute(attributes, path...)
	if !ok {
		return false
	}
	asBool, _ := value.(bool)
	return asBool
}

// locationAttribute walks a nested attributes map, returning the value at path
// when every level is present.
func locationAttribute(attributes map[string]interface{}, path ...string) (interface{}, bool) {
	var current interface{} = attributes
	for _, key := range path {
		asMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = asMap[key]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
