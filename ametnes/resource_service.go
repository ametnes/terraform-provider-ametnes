package ametnes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceCreate,
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		Schema: map[string]*schema.Schema{
			"items": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"kind": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"location": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"network": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// ...
	var diags diag.Diagnostics
	return diags
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// ...
	// client := &http.Client{Timeout: 10 * time.Second}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%s/resources/%s", "https://api-test.cloud.ametnes.com/v1"), nil)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	// client, err := NewClient(&host, &username, &token)

	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	// projects, err := client.GetProjects()
	// if err != nil {
	// 	return diag.FromErr(err)
	// }
	// if len(projects) == 0 {
	// 	return diag.FromErr(err)
	// }
	// project := projects[0]

	// resource := Resource{
	// 	Name:     "Test Resource 6",
	// 	Project:  project.Id,
	// 	Account:  project.Account,
	// 	Kind:     "service/mysql:8.0",
	// 	Location: "gcp/europe-west2",
	// }

	// n_resource, err := client.CreateResource(resource)

	// r, err := client.Do(req)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }
	// defer r.Body.Close()

	// var kinds map[string]interface{}
	// err = json.NewDecoder(r.Body).Decode(&kinds)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	// m_kinds := make([]interface{}, len(kinds["results"].([]interface{})))
	// for idx, __kind := range kinds["results"].([]interface{}) {
	// 	kind := __kind.(map[string]interface{})
	// 	_kind := make(map[string]interface{})
	// 	_kind["id"] = kind["id"]
	// 	_kind["name"] = kind["name"]
	// 	_kind["type"] = kind["type"]
	// 	_kind["enabled"] = kind["enabled"]
	// 	_kind["release"] = kind["release"]
	// 	_kind["locations"] = kind["locations"]
	// 	_kind["kind"] = kind["kind"]

	// 	limits := make([]interface{}, 1)
	// 	limits[0] = kind["limits"]
	// 	_kind["limits"] = limits

	// 	backups := make([]interface{}, 1)
	// 	backups[0] = kind["backups"]
	// 	_kind["backups"] = backups

	// 	_kind["tools"] = nil

	// 	m_kinds[idx] = _kind
	// }

	// if err := d.Set("kinds", m_kinds); err != nil {
	// 	return diag.FromErr(err)
	// }

	return diags
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// ...
	var diags diag.Diagnostics
	return diags
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// ...
	var diags diag.Diagnostics
	return diags
}
