package ametnes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Unit tests using mock HTTP server (no real API calls)
func TestClient_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	projects, err := client.GetProjects()
	assert.NoError(t, err)
	assert.NotNil(t, projects)
	assert.Greater(t, len(projects), 0)
	assert.Equal(t, "Test Project 1", projects[0].Name)
	assert.Equal(t, 1, projects[0].Id)
}

func TestClient_GetProjects_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	projects, err := client.GetProjects()
	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, 1, projects[0].Id)
	assert.Equal(t, 2, projects[1].Id)
}

func TestClient_CreateProject_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	project := Project{
		Name:        "New Project",
		Description: "Test Description",
	}

	created, err := client.CreateProject(project)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, 999, created.Id)
	assert.Equal(t, "New Project", created.Name)
	assert.True(t, created.Enabled)
}

func TestClient_GetResources_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	project := &Project{Id: 1}
	resources, err := client.GetResources(project)
	assert.NoError(t, err)
	assert.NotNil(t, resources)
	assert.Greater(t, len(resources), 0)
	assert.Equal(t, "Test Resource", resources[0].Name)
}

func TestClient_GetResource_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	resource, err := client.GetResource(1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resource)
	assert.Equal(t, "Test Resource", resource.Name)
	assert.Equal(t, "ONLINE", resource.Status)
	assert.NotNil(t, resource.Spec.Connection)
	assert.Equal(t, "test.example.com", resource.Spec.Connection.Host)
}

func TestClient_CreateResource_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	resource := Resource{
		Project:  1,
		Kind:     "service/mysql:8.0",
		Location: "gcp/europe-west2",
		Name:     "New Resource",
		Spec: Spec{
			Components: map[string]interface{}{
				"cpu":     1,
				"memory":  1,
				"storage": 1,
			},
			Nodes: 1,
		},
	}

	created, err := client.CreateResource(resource)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, 999, created.Id)
	assert.Equal(t, "INIT", created.Status)
}

func TestClient_UpdateResource_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	resource := Resource{
		Id:      1,
		Project: 1,
		Name:    "Updated Resource",
		Spec: Spec{
			Config: map[string]interface{}{
				"new.config": "new-value",
			},
		},
	}

	updated, err := client.UpdateResource(resource)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "ONLINE", updated.Status)
}

func TestClient_DeleteResource_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	err := client.DeleteResource(Resource{
		Id:      1,
		Project: 1,
	})
	assert.NoError(t, err)
}

func TestClient_GetLocations_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	locations, err := client.GetLocations()
	assert.NoError(t, err)
	assert.NotNil(t, locations)
	assert.Len(t, locations, 2)
	assert.Equal(t, "gcp.europe-west2", locations[0].Id)
	assert.Equal(t, "London, U.K.", locations[0].Name)
}

func TestClient_CreateLocation_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	location := Location{
		Name:     "Test Location",
		Location: "test/region",
	}

	created, err := client.CreateLocation(location)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, "test.location", created.Id)
	assert.True(t, created.Enabled)
}

func TestClient_DeleteLocation_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	err := client.DeleteLocation(Location{Id: "test.location"})
	assert.NoError(t, err)
}

