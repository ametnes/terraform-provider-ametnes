package ametnes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocationSchemaDefaults_Unit(t *testing.T) {
	rsrc := resourceLocation()
	require.NoError(t, Provider().InternalValidate())

	d := schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"name": "Acme",
		"code": "USE1",
	})

	// Both the network and telemetry blocks are optional and computed.
	assert.Empty(t, d.Get("network").([]interface{}))
	assert.Empty(t, d.Get("telemetry").([]interface{}))
}

func TestExpandNetworkDefaults_Unit(t *testing.T) {
	// A missing network block resolves to the platform defaults: private only,
	// with private/location defaults.
	network := expandNetwork(nil)
	assert.Equal(t, false, network["public"])
	assert.Equal(t, true, network["private"])
	assert.Equal(t, DefaultLoadBalancerVisibility, network["visibility"])
	assert.Equal(t, DefaultLoadBalancerScope, network["scope"])
}

func TestExpandLocationAttributes_Unit(t *testing.T) {
	rsrc := resourceLocation()
	d := schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"name": "Acme",
		"code": "USE1",
		"network": []interface{}{
			map[string]interface{}{
				"public":     true,
				"private":    true,
				"visibility": "public",
				"scope":      "project",
			},
		},
		"telemetry": []interface{}{
			map[string]interface{}{
				"resources": []interface{}{
					map[string]interface{}{"logs": true, "metrics": false},
				},
				"system": []interface{}{
					map[string]interface{}{"logs": false, "metrics": true},
				},
			},
		},
	})

	attrs := expandLocationAttributes(d)

	lbCaps := locationNested(attrs, "capabilities", "network", "load_balancer")
	assert.Equal(t, true, lbCaps["public"])
	assert.Equal(t, true, lbCaps["private"])

	lbDefaults := locationNested(attrs, "defaults", "network", "load_balancer")
	assert.Equal(t, "public", lbDefaults["visibility"])
	assert.Equal(t, "project", lbDefaults["scope"])

	telemetryResources := locationNested(attrs, "telemetry", "resources")
	assert.Equal(t, true, locationNested(telemetryResources, "logs")["enabled"])
	assert.Equal(t, false, locationNested(telemetryResources, "metrics")["enabled"])

	telemetrySystem := locationNested(attrs, "telemetry", "system")
	assert.Equal(t, false, locationNested(telemetrySystem, "logs")["enabled"])
	assert.Equal(t, true, locationNested(telemetrySystem, "metrics")["enabled"])
}

func TestExpandTelemetryDefaultsDisabled_Unit(t *testing.T) {
	// An empty telemetry block still produces a complete, disabled payload.
	attributes := expandTelemetry([]interface{}{})
	assert.Equal(t, false, locationNested(attributes, "resources", "logs")["enabled"])
	assert.Equal(t, false, locationNested(attributes, "resources", "metrics")["enabled"])
	assert.Equal(t, false, locationNested(attributes, "system", "logs")["enabled"])
	assert.Equal(t, false, locationNested(attributes, "system", "metrics")["enabled"])
}

func TestSetLocationAttributes_Unit(t *testing.T) {
	rsrc := resourceLocation()
	d := schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"name": "Acme",
		"code": "USE1",
	})

	attributes := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"network": map[string]interface{}{
				"load_balancer": map[string]interface{}{"public": false, "private": true},
			},
		},
		"defaults": map[string]interface{}{
			"network": map[string]interface{}{
				"load_balancer": map[string]interface{}{"visibility": "private", "scope": "project"},
			},
		},
		"telemetry": map[string]interface{}{
			"resources": map[string]interface{}{
				"logs":    map[string]interface{}{"enabled": true},
				"metrics": map[string]interface{}{"enabled": false},
			},
			"system": map[string]interface{}{
				"logs":    map[string]interface{}{"enabled": false},
				"metrics": map[string]interface{}{"enabled": true},
			},
		},
	}

	setLocationAttributes(d, attributes)

	network := d.Get("network").([]interface{})
	require.Len(t, network, 1)
	block := network[0].(map[string]interface{})
	assert.Equal(t, false, block["public"])
	assert.Equal(t, true, block["private"])
	assert.Equal(t, "private", block["visibility"])
	assert.Equal(t, "project", block["scope"])

	assert.Equal(t, true, telemetryBlockValue(t, d, "resources", "logs"))
	assert.Equal(t, false, telemetryBlockValue(t, d, "resources", "metrics"))
	assert.Equal(t, false, telemetryBlockValue(t, d, "system", "logs"))
	assert.Equal(t, true, telemetryBlockValue(t, d, "system", "metrics"))
}

func TestNetworkSpecConfig_Unit(t *testing.T) {
	rsrc := resourceNetwork()

	// Unset visibility/scope are left out so the location defaults apply.
	d := schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"project":  "1",
		"name":     "net",
		"location": "gcp/europe-west2",
	})
	config := networkSpecConfig(d)
	_, hasVisibility := config["visibility"]
	_, hasScope := config["scope"]
	assert.False(t, hasVisibility, "visibility should be omitted when unset")
	assert.False(t, hasScope, "scope should be omitted when unset")

	// Explicit values and advanced config keys are carried through.
	d = schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"project":    "1",
		"name":       "net",
		"location":   "gcp/europe-west2",
		"visibility": "public",
		"scope":      "project",
		"config":     map[string]interface{}{"services_per_lb": "3"},
	})
	config = networkSpecConfig(d)
	assert.Equal(t, "public", config["visibility"])
	assert.Equal(t, "project", config["scope"])
	assert.Equal(t, "3", config["services_per_lb"])
}

func TestNetworkVisibilityScopeValidation_Unit(t *testing.T) {
	rsrc := resourceNetwork()
	require.NoError(t, Provider().InternalValidate())

	_, errs := rsrc.Schema["visibility"].ValidateFunc("public", "visibility")
	assert.Empty(t, errs)
	_, errs = rsrc.Schema["visibility"].ValidateFunc("bogus", "visibility")
	assert.NotEmpty(t, errs)

	_, errs = rsrc.Schema["scope"].ValidateFunc("project", "scope")
	assert.Empty(t, errs)
	_, errs = rsrc.Schema["scope"].ValidateFunc("bogus", "scope")
	assert.NotEmpty(t, errs)
}

func TestLocationValidation_Unit(t *testing.T) {
	rsrc := resourceLocation()
	require.NoError(t, Provider().InternalValidate())

	networkSchema := rsrc.Schema["network"].Elem.(*schema.Resource).Schema

	_, errs := networkSchema["visibility"].ValidateFunc("bogus", "visibility")
	assert.NotEmpty(t, errs)
	_, errs = networkSchema["visibility"].ValidateFunc("private", "visibility")
	assert.Empty(t, errs)

	_, errs = networkSchema["scope"].ValidateFunc("bogus", "scope")
	assert.NotEmpty(t, errs)
	_, errs = networkSchema["scope"].ValidateFunc("location", "scope")
	assert.Empty(t, errs)
}

func TestUpdateLocation_Unit(t *testing.T) {
	var received Location
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "PUT" && r.URL.Path == "/metadata/locations/test-location":
			json.NewDecoder(r.Body).Decode(&received)
			received.Status = "ONLINE"
			json.NewEncoder(w).Encode(received)
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
		}
	}))
	defer server.Close()

	username := "test-user"
	client, err := NewClient(server.URL, Token{Type: Basic, Username: &username, Token: "test-token"})
	require.NoError(t, err)

	updated, err := client.UpdateLocation(Location{
		Id:       "test-location",
		Name:     "Acme",
		Location: "USE1",
		Attributes: map[string]interface{}{
			"defaults": map[string]interface{}{
				"network": map[string]interface{}{
					"load_balancer": map[string]interface{}{"visibility": "private", "scope": "location"},
				},
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "test-location", received.Id)
	assert.Equal(t, "Acme", received.Name)
	assert.Equal(t, "private", locationNested(received.Attributes, "defaults", "network", "load_balancer")["visibility"])
	assert.Equal(t, "ONLINE", updated.Status)
}

func TestLocationUpdateContext_Unit(t *testing.T) {
	putCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "PUT" && r.URL.Path == "/metadata/locations/loc-1":
			putCalled = true
			var location Location
			json.NewDecoder(r.Body).Decode(&location)
			location.Status = "ONLINE"
			json.NewEncoder(w).Encode(location)
		case r.Method == "GET" && r.URL.Path == "/metadata/locations":
			json.NewEncoder(w).Encode(Locations{
				Count: 1,
				Items: []Location{
					{Id: "loc-1", Name: "Acme", Location: "USE1", Enabled: true, Status: "ONLINE"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
		}
	}))
	defer server.Close()

	username := "test-user"
	client, err := NewClient(server.URL, Token{Type: Basic, Username: &username, Token: "test-token"})
	require.NoError(t, err)

	rsrc := resourceLocation()
	d := schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"name": "Acme",
		"code": "USE1",
		"network": []interface{}{
			map[string]interface{}{"public": false, "private": true, "visibility": "private", "scope": "location"},
		},
	})
	d.SetId("loc-1")
	// Simulate an in-place change so the update context issues a PUT.
	d.Set("network", []interface{}{
		map[string]interface{}{"public": true, "private": true, "visibility": "public", "scope": "location"},
	})

	diags := resourceLocationUpdate(context.Background(), d, client)
	assert.False(t, diags.HasError(), "expected in-place location update to succeed")
	assert.True(t, putCalled, "expected the location update to call the API")
}

// telemetryBlockValue reads a telemetry feature from the telemetry block in
// state.
func telemetryBlockValue(t *testing.T, d *schema.ResourceData, scope, feature string) bool {
	t.Helper()
	telemetry := d.Get("telemetry").([]interface{})
	require.Len(t, telemetry, 1)
	block := telemetry[0].(map[string]interface{})
	scopes := block[scope].([]interface{})
	require.Len(t, scopes, 1)
	scopeMap := scopes[0].(map[string]interface{})
	value, _ := scopeMap[feature].(bool)
	return value
}

// locationNested returns the nested map at path for assertion convenience.
func locationNested(root map[string]interface{}, path ...string) map[string]interface{} {
	value, ok := locationAttribute(root, path...)
	if !ok {
		return map[string]interface{}{}
	}
	asMap, _ := value.(map[string]interface{})
	return asMap
}
