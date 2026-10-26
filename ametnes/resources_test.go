package ametnes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Existing tests...

func TestGetResources_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	projects, err := client.GetProjects()
	assert.NoError(t, err)
	assert.Greater(t, len(projects), 0)

	project := projects[0]

	resources, err := client.GetResources(&project)
	assert.NoError(t, err)
	assert.NotNil(t, resources)
}

func TestCreateAndGetResource_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	projects, err := client.GetProjects()
	assert.NoError(t, err)
	assert.Greater(t, len(projects), 0)

	project := projects[0]

	spec := Spec{}
	components := make(map[string]interface{})
	components["storage"] = 1

	spec.Components = components
	spec.Nodes = 1

	resource := Resource{
		Name:     "Test Resource",
		Project:  project.Id,
		Account:  project.Account,
		Kind:     "service/mysql:8.0",
		Location: "gcp/europe-west2",
		Spec:     spec,
		Product:  DefaultProductCode,
	}

	n_resource, err := client.CreateResource(resource)
	assert.NoError(t, err)
	assert.NotNil(t, n_resource)
	assert.Equal(t, "INIT", n_resource.Status)

	gresource, err2 := client.GetResource(project.Id, n_resource.Id)
	assert.NoError(t, err2)
	assert.Equal(t, gresource.Kind, resource.Kind)
}

// New test for network field immutability
func TestResourceService_NetworkFieldImmutability(t *testing.T) {
	// Setup dummy schema.ResourceData
	resource := resourceService()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		"project":  "1",
		"name":     "test-service",
		"kind":     "mysql:8.0",
		"location": "gcp/europe-west2",
		"network":  "123",
	})

	// Simulate a change to the "network" field
	d.SetId("1/1")
	d.Set("network", "123")
	d.Set("network", "456") // simulate change

	// Simulate HasChange("network") returning true by using a custom ResourceData wrapper
	// Since schema.ResourceData does not expose SetPartial or HasChange, we need to call the UpdateContext
	// and expect it to fail if the implementation is correct. This is a limitation of the SDK's test helpers.
	// In a real test, you would use a mocking framework or integration test.

	// Call the UpdateContext function directly
	updateFunc := resource.UpdateContext
	diags := updateFunc(context.Background(), d, nil)

	require.True(t, diags.HasError(), "Expected error when changing 'network' field")
	require.Contains(t, diags[0].Summary, "Changing the 'network' field is not permitted")
}

func TestResourceService_KindIsUpdatable(t *testing.T) {
	resource := resourceService()

	kind := resource.Schema["kind"]
	require.NotNil(t, kind)
	require.False(t, kind.ForceNew, "kind must not force replacement: the control plane upgrades services in place")
}

func TestResourceService_KindUpgradeSendsPrefixedKind(t *testing.T) {
	var putResource Resource
	appliedKind := "service/grafana:8.5"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPut:
			if err := json.NewDecoder(r.Body).Decode(&putResource); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			appliedKind = putResource.Kind
			putResource.Status = "READY"
			_ = json.NewEncoder(w).Encode(putResource)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(Resource{
				Id:          1,
				Project:     1,
				Account:     100,
				Kind:        appliedKind,
				Location:    "gcp/europe-west2",
				Name:        "test-service",
				Description: "test service",
				Status:      "READY",
				Network:     2,
				Spec: Spec{
					Components: map[string]interface{}{"storage": 10},
					Nodes:      1,
					Config:     map[string]interface{}{"architecture": "Starter"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	username := "test-user"
	client, err := NewClient(server.URL, Token{Type: Basic, Username: &username, Token: "test-token"})
	require.NoError(t, err)

	serviceResource := resourceService()
	state := &terraform.InstanceState{
		ID: "1/1",
		Attributes: map[string]string{
			"project":             "1",
			"name":                "test-service",
			"kind":                "grafana:8.5",
			"location":            "gcp/europe-west2",
			"network":             "1/2",
			"capacity.#":          "1",
			"capacity.0.storage":  "10",
			"config.%":            "1",
			"config.architecture": "Starter",
		},
	}
	config := terraform.NewResourceConfigRaw(map[string]interface{}{
		"project":  "1",
		"name":     "test-service",
		"kind":     "grafana:9.3",
		"location": "gcp/europe-west2",
		"network":  "1/2",
		"capacity": []interface{}{
			map[string]interface{}{"storage": 10},
		},
		"config": map[string]interface{}{"architecture": "Starter"},
	})

	ctx := context.Background()
	diff, err := serviceResource.Diff(ctx, state, config, client)
	require.NoError(t, err)
	require.NotNil(t, diff)
	require.False(t, diff.RequiresNew(), "changing kind must not require replacement")
	require.Contains(t, diff.Attributes, "kind")

	newState, diags := serviceResource.Apply(ctx, state, diff, client)
	require.False(t, diags.HasError(), "update failed: %v", diags)

	assert.Equal(t, "service/grafana:9.3", putResource.Kind)
	require.NotNil(t, newState)
	assert.Equal(t, "grafana:9.3", newState.Attributes["kind"])
}

func TestResourceService_Timeouts(t *testing.T) {
	resource := resourceService()

	require.NotNil(t, resource.Timeouts)
	require.NotNil(t, resource.Timeouts.Create)
	require.NotNil(t, resource.Timeouts.Update)
	require.NotNil(t, resource.Timeouts.Delete)

	assert.Equal(t, 60*time.Minute, *resource.Timeouts.Create)
	assert.Equal(t, 45*time.Minute, *resource.Timeouts.Update)
	assert.Equal(t, 10*time.Minute, *resource.Timeouts.Delete)
}

func TestResourceNetwork_Timeouts(t *testing.T) {
	resource := resourceNetwork()

	require.NotNil(t, resource.Timeouts)
	require.NotNil(t, resource.Timeouts.Create)
	require.NotNil(t, resource.Timeouts.Update)
	require.NotNil(t, resource.Timeouts.Delete)

	assert.Equal(t, 15*time.Minute, *resource.Timeouts.Create)
	assert.Equal(t, 15*time.Minute, *resource.Timeouts.Update)
	assert.Equal(t, 10*time.Minute, *resource.Timeouts.Delete)
}
