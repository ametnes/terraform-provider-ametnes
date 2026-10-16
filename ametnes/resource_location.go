package ametnes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Load balancer capability/default and telemetry values shared with the
// Ametnes Cloud console. See api.cloud.ametnes.com
// (bin/cloud/services/metadata.py and bin/cloud/services/util.py) for the server
// side contract; the console edits both through a single
// `PUT metadata/locations` request.
const (
	DefaultLoadBalancerVisibility = "private"
	DefaultLoadBalancerScope      = "location"
)

func resourceLocation() *schema.Resource {
	return &schema.Resource{
		Description: `
Creates and manages an Ametnes Data Service location.

A location is a dedicated Kubernetes cluster with the Ametnes Cloud Agent
installed. Its configuration controls how load balancers are provisioned for the
services created in it, and which telemetry the agent collects:

- The ` + "`network`" + ` block sets the load balancer policy: the visibilities
  services are allowed to request (` + "`public`" + `, ` + "`private`" + `) and the defaults
  used when a service or network does not choose (` + "`visibility`" + `, ` + "`scope`" + `).
  Default visibility is ` + "`private`" + `; default scope is ` + "`location`" + ` (cluster wide).
- The ` + "`telemetry`" + ` block switches the resource and system ` + "`logs`" + ` and
  ` + "`metrics`" + ` the agent collects. Logs and metrics require a Premium or
  Enterprise plan.

**On-premises:** load balancer visibility is irrelevant because every load
balancer is provisioned inside your own environment.

These settings are updated in place; the location identity (` + "`name`" + `, ` + "`code`" + `,
` + "`description`" + `) is immutable.
`,
		ReadContext:   dataSourceLocationsRead,
		CreateContext: resourceLocationCreate,
		UpdateContext: resourceLocationUpdate,
		DeleteContext: resourceLocationDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the data service location.",
			},
			"code": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "An easy identifier of the data service location. A combination of the `name` and `code` must be unique.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Description of the location.",
			},
			"network": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Load balancer policy for services in this location: the visibilities they are allowed to request (`public`, `private`) and the defaults used when a service or network does not choose (`visibility`, `scope`). When omitted, the platform defaults apply.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"public": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this location allows **public** (internet facing) load balancers. Capabilities are the visibilities services at the location are allowed to request. Defaults to `false`. On-premises: visibility is irrelevant, because every load balancer is provisioned inside your own environment.",
						},
						"private": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether this location allows **private** (internal) load balancers. Defaults to `true`. On-premises: visibility is irrelevant, because every load balancer is provisioned inside your own environment.",
						},
						"visibility": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      DefaultLoadBalancerVisibility,
							ValidateFunc: validation.StringInSlice([]string{VisibilityPublic, VisibilityPrivate}, false),
							Description:  "Default visibility for load balancers created in this location when a service or network does not choose one. `private` (the default) provisions an internal load balancer; `public` provisions an internet facing one. Must be one of the allowed capabilities. On-premises: visibility is irrelevant, because every load balancer is provisioned inside your own environment.",
						},
						"scope": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      DefaultLoadBalancerScope,
							ValidateFunc: validation.StringInSlice([]string{ScopeLocation, ScopeProject}, false),
							Description:  "Default scope for load balancers created in this location when a service or network does not choose one. `location` (the default) makes them cluster wide so every project in the location can attach to them; `project` scopes them to the owning project (namespace) only.",
						},
					},
				},
			},
			"telemetry": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Telemetry the location's agent collects. Requires a Premium or Enterprise plan. When omitted, all telemetry is disabled.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": telemetryScopeSchema("logs and metrics for the data services (resources) running at this location."),
						"system":    telemetryScopeSchema("logs and metrics for the location's system components."),
					},
				},
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "`true` if this location is enabled and can have resources created in it.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
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
		Attributes:  expandLocationAttributes(d),
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(createdLocation.Id)

	return dataSourceLocationsRead(ctx, d, m)

}

func resourceLocationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	fields := []string{
		"network",
		"telemetry",
	}

	hasChanges := false
	for _, field := range fields {
		if d.HasChange(field) {
			hasChanges = true
			break
		}
	}

	if hasChanges {
		locationDescription := ""
		if description, ok := d.GetOk("description"); ok {
			locationDescription = description.(string)
		}

		_, err := client.UpdateLocation(Location{
			Id:          d.Id(),
			Name:        d.Get("name").(string),
			Location:    d.Get("code").(string),
			Description: locationDescription,
			Attributes:  expandLocationAttributes(d),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

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

// expandLocationAttributes builds the location attributes payload understood by
// the cloud API. It mirrors the structure the Ametnes Cloud console sends when
// editing a location: load balancer capabilities and defaults plus telemetry.
func expandLocationAttributes(d *schema.ResourceData) map[string]interface{} {
	network := expandNetwork(d.Get("network").([]interface{}))
	return map[string]interface{}{
		"capabilities": map[string]interface{}{
			"network": map[string]interface{}{
				"load_balancer": map[string]interface{}{
					"public":  network["public"],
					"private": network["private"],
				},
			},
		},
		"defaults": map[string]interface{}{
			"network": map[string]interface{}{
				"load_balancer": map[string]interface{}{
					"visibility": network["visibility"],
					"scope":      network["scope"],
				},
			},
		},
		"telemetry": expandTelemetry(d.Get("telemetry").([]interface{})),
	}
}

// expandNetwork resolves the `network` block into capability and default values,
// falling back to the platform defaults when the block (or a field) is unset.
func expandNetwork(networkList []interface{}) map[string]interface{} {
	public := false
	private := true
	visibility := DefaultLoadBalancerVisibility
	scope := DefaultLoadBalancerScope
	if len(networkList) > 0 && networkList[0] != nil {
		block := networkList[0].(map[string]interface{})
		if v, ok := block["public"].(bool); ok {
			public = v
		}
		if v, ok := block["private"].(bool); ok {
			private = v
		}
		if v, ok := block["visibility"].(string); ok && v != "" {
			visibility = v
		}
		if v, ok := block["scope"].(string); ok && v != "" {
			scope = v
		}
	}
	return map[string]interface{}{
		"public":     public,
		"private":    private,
		"visibility": visibility,
		"scope":      scope,
	}
}

// telemetryScopeSchema is the shared schema for the `resources` and `system`
// telemetry scopes. Leaves default to disabled so a partially specified block
// still produces a complete payload.
func telemetryScopeSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Description: "Telemetry for " + description,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"logs": {
					Type:        schema.TypeBool,
					Optional:    true,
					Computed:    true,
					Description: "Whether logs are collected. Requires a Premium or Enterprise plan.",
				},
				"metrics": {
					Type:        schema.TypeBool,
					Optional:    true,
					Computed:    true,
					Description: "Whether metrics are collected. Requires a Premium or Enterprise plan.",
				},
			},
		},
	}
}

// expandTelemetry converts the telemetry block into the cloud API payload,
// defaulting every scope and feature to disabled when unset.
func expandTelemetry(telemetryList []interface{}) map[string]interface{} {
	scopes := map[string]interface{}{}
	if len(telemetryList) > 0 && telemetryList[0] != nil {
		telemetry := telemetryList[0].(map[string]interface{})
		scopes["resources"] = expandTelemetryScope(telemetry["resources"])
		scopes["system"] = expandTelemetryScope(telemetry["system"])
	} else {
		scopes["resources"] = expandTelemetryScope(nil)
		scopes["system"] = expandTelemetryScope(nil)
	}
	return scopes
}

func expandTelemetryScope(value interface{}) map[string]interface{} {
	logs := false
	metrics := false
	if list, ok := value.([]interface{}); ok && len(list) > 0 && list[0] != nil {
		scope := list[0].(map[string]interface{})
		if v, ok := scope["logs"].(bool); ok {
			logs = v
		}
		if v, ok := scope["metrics"].(bool); ok {
			metrics = v
		}
	}
	return map[string]interface{}{
		"logs":    map[string]interface{}{"enabled": logs},
		"metrics": map[string]interface{}{"enabled": metrics},
	}
}
