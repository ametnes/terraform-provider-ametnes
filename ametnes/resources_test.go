package ametnes

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	components["cpu"] = 1
	components["memory"] = 1
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
		"nodes":    1,
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
