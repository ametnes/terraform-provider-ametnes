package ametnes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKinds() *schema.Resource {
	return &schema.Resource{
		Description: "Read resource kinds metadata",
		ReadContext: dataSourceKindsRead,
		Schema: map[string]*schema.Schema{
			"kinds": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
							Description: "Unique backend identifier for the resource kind. Example: `e834def`.",
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
							Description: "Name of the resource kind.",
						},
						"kind": {
							Type:     schema.TypeString,
							Computed: true,
							Description: "Kind code to be used when creating resources. Example `service/grafana:8.3`",
						},
						"locations": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem:     schema.TypeString,
						},
						"backups": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
							Description: "`true` if backups are enabled for this resource kind",
						},
						"tools": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"limits": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"nodes": {
										MaxItems: 1,
										Type:     schema.TypeList,
										Required: true,
										Elem:     schema.TypeInt,
									},
								},
							},
							Description: "Capacity limits for ths resource kind. These include `memory`, `cpu` and `storage` as well as the number of `nodes` that can be scaled to.",
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
							Description: "Global kind code to be used only for information. Example `service/grafana`",
						},
						"release": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
							Description: "`true` if resourcees of this kind can be created.",
						},
					},
				},
			},
		},
	}
}

func dataSourceKindsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := &http.Client{Timeout: 10 * time.Second}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/metadata/resources/kinds", "https://api-test.cloud.ametnes.com/v1"), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	var kinds map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&kinds)
	if err != nil {
		return diag.FromErr(err)
	}

	m_kinds := make([]interface{}, len(kinds["results"].([]interface{})))
	for idx, __kind := range kinds["results"].([]interface{}) {
		kind := __kind.(map[string]interface{})
		_kind := make(map[string]interface{})
		_kind["id"] = kind["id"]
		_kind["name"] = kind["name"]
		_kind["type"] = kind["type"]
		_kind["enabled"] = kind["enabled"]
		_kind["release"] = kind["release"]
		_kind["locations"] = kind["locations"]
		_kind["kind"] = kind["kind"]

		limits := make([]interface{}, 1)
		limits[0] = kind["limits"]
		_kind["limits"] = limits

		backups := make([]interface{}, 1)
		backups[0] = kind["backups"]
		_kind["backups"] = backups

		_kind["tools"] = nil

		m_kinds[idx] = _kind
	}

	if err := d.Set("kinds", m_kinds); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
