package ametnes

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
# curl -k -X GET "https://cloud.ametnes.com/api/c/v1/metadata/locations"
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

func TestKindData_Unit(t *testing.T) {
	var gotPath, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count": 1,
			"results": []interface{}{
				map[string]interface{}{
					"id":        "e834def",
					"name":      "Redis 8.10",
					"kind":      "redis:8.10",
					"type":      "service/redis",
					"release":   "8.10",
					"locations": []interface{}{"gcp.europe-west2"},
					"limits": map[string]interface{}{
						"nodes": []interface{}{1, 3},
						"storage": map[string]interface{}{
							"min":       map[string]interface{}{"units": 10, "units_uom": "GB"},
							"units":     4000,
							"units_uom": "GB",
						},
						"memory": map[string]interface{}{
							"min":       map[string]interface{}{"units": 2, "units_uom": "GB"},
							"units":     128,
							"units_uom": "GB",
						},
						"cpu": map[string]interface{}{
							"min":   map[string]interface{}{"units": 1},
							"units": 48,
						},
					},
					"backups": map[string]interface{}{"enabled": true, "quiesce": true},
				},
			},
		})
	}))
	defer server.Close()

	username := "test-user"
	client, err := NewClient(server.URL, Token{Type: Basic, Username: &username, Token: "test-token"})
	require.NoError(t, err)

	d := schema.TestResourceDataRaw(t, dataSourceKinds().Schema, nil)
	diags := dataSourceKindsRead(context.Background(), d, client)
	require.False(t, diags.HasError(), "read failed: %v", diags)

	assert.Equal(t, "/metadata/resources/kinds", gotPath, "the data source must query the configured provider host")
	assert.Equal(t, client.AuthHeader, gotAuth, "the data source must authenticate with the provider credentials")
	kinds := d.Get("kinds").([]interface{})
	require.Len(t, kinds, 1)
	assert.Equal(t, true, kinds[0].(map[string]interface{})["enabled"], "catalogue kinds are returned only when enabled")
}

func TestKindData(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Set TF_ACC to run acceptance tests against the live API")
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	host := os.Getenv("AMETNES_HOST")
	if host == "" {
		host = AMETNES_HOST
	}
	username := os.Getenv("AMETNES_USERNAME")
	client, err := NewClient(host, Token{Type: Basic, Username: &username, Token: os.Getenv("AMETNES_TOKEN")})
	require.NoError(t, err)

	resourceData := schema.TestResourceDataRaw(t, dataSourceKinds().Schema, nil)
	diag := dataSourceKindsRead(context.TODO(), resourceData, client)
	if diag.HasError() {
		// Check if it's an authentication or API error
		for _, d := range diag {
			errorMsg := d.Summary + " " + d.Detail
			if strings.Contains(errorMsg, "401") ||
				strings.Contains(errorMsg, "Unauthorized") {
				t.Skipf("Skipping test due to authentication error: %s - %s", d.Summary, d.Detail)
				return
			}
		}
		t.Fail()
	}
}
