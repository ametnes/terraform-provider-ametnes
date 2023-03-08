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

/*
# curl -k -X GET "https://api.cloud.ametnes.com/v1/metadata/resources/kinds"
{
  "count": 7,
  "results": [
    {
      "backups": {
        "enabled": false
      },
      "id": "net.privatelink10",
      "kind": "network/privatelink:1.0",
      "limits": null,
      "locations": [
        "aws/eu-west-2"
      ],
      "name": "Private Link Network",
      "release": null,
      "tools": null,
      "type": "network/privatelink"
    },
    {
      "backups": {
        "enabled": true
      },
      "id": "svc.mysql80",
      "kind": "service/mysql:8.0",
      "limits": null,
      "locations": [
        "gcp/europe-west2",
        "aws/eu-west-2"
      ],
      "name": "MySql 8.0 Service",
      "release": null,
      "tools": null,
      "type": "service/mysql"
    },
    {
      "backups": {
        "enabled": false
      },
      "id": "svc.neo4j42",
      "kind": "service/neo4j:4.2",
      "limits": {
        "nodes": [
          1,
          3
        ]
      },
      "locations": [
        "gcp/europe-west2",
        "aws/eu-west-2"
      ],
      "name": "Neo4J 4.2 Service",
      "release": null,
      "tools": null,
      "type": "service/neo4j"
    },
    {
      "backups": {
        "enabled": true
      },
      "id": "svc.postgres119",
      "kind": "service/postgres:11.9",
      "limits": null,
      "locations": [
        "gcp/europe-west2",
        "aws/eu-west-2"
      ],
      "name": "Postgres 11.9 Service",
      "release": null,
      "tools": null,
      "type": "service/postgres"
    },
    {
      "backups": {
        "enabled": true
      },
      "id": "svc.postgres124",
      "kind": "service/postgres:12.4",
      "limits": null,
      "locations": [
        "gcp/europe-west2",
        "aws/eu-west-2"
      ],
      "name": "Postgres 12.4 Service",
      "release": null,
      "tools": null,
      "type": "service/postgres"
    },
    {
      "backups": {
        "enabled": true
      },
      "id": "svc.postgres130",
      "kind": "service/postgres:13.0",
      "limits": null,
      "locations": [
        "gcp/europe-west2",
        "aws/eu-west-2"
      ],
      "name": "Postgres 13.0 Service",
      "release": null,
      "tools": null,
      "type": "service/postgres"
    },
    {
      "backups": {
        "enabled": false
      },
      "id": "svc.redis62",
      "kind": "service/redis:6.2",
      "limits": null,
      "locations": [
        "gcp/europe-west2",
        "aws/eu-west-2"
      ],
      "name": "Redis 6.2 Service",
      "release": "beta",
      "tools": null,
      "type": "service/redis"
    }
  ]
}
*/

func dataSourceKinds() *schema.Resource {
	return &schema.Resource{
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
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kind": {
							Type:     schema.TypeString,
							Computed: true,
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
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"release": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
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
