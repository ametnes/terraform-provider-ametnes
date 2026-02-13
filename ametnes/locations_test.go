package ametnes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocations_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	locations, err := client.GetLocations()
	assert.NoError(t, err)
	assert.NotNil(t, locations)
	assert.Greater(t, len(locations), 0)

	// Verify the structure of at least one location
	location := locations[0]
	assert.NotEmpty(t, location.Id)
	assert.NotEmpty(t, location.Name)
	assert.NotEmpty(t, location.Location)
}

func TestCreateAndDeleteLocation_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	newLocation := Location{
		Name:        "Test Location",
		Description: "Test location created by unit tests",
		Enabled:     true,
		Location:    "test-region",
	}

	createdLocation, err := client.CreateLocation(newLocation)
	assert.NoError(t, err)
	assert.NotNil(t, createdLocation)
	assert.NotEmpty(t, createdLocation.Id)
	assert.Equal(t, newLocation.Name, createdLocation.Name)
	assert.Equal(t, newLocation.Description, createdLocation.Description)
	assert.Equal(t, newLocation.Enabled, createdLocation.Enabled)
	assert.Equal(t, "test.location", createdLocation.Id)

	// Test deletion
	err = client.DeleteLocation(*createdLocation)
	assert.NoError(t, err)
}

func TestGetLocationsReturnsValidStructure_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	locations, err := client.GetLocations()
	assert.NoError(t, err)

	// Test that locations have the expected fields
	for _, location := range locations {
		assert.NotEmpty(t, location.Id, "Location ID should not be empty")
		assert.NotEmpty(t, location.Name, "Location name should not be empty")
		assert.NotEmpty(t, location.Location, "Location code should not be empty")

		// Status should be one of the expected values if present
		if location.Status != "" {
			assert.Contains(t, []string{"ONLINE", "OFFLINE", "MAINTENANCE"}, location.Status,
				"Location status should be one of ONLINE, OFFLINE, or MAINTENANCE")
		}
	}
}

func TestLocationOperationsWithInvalidData_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	// Test creating location with minimal data
	minimalLocation := Location{
		Name:     "Minimal Test Location",
		Location: "minimal-test-region",
	}

	createdLocation, err := client.CreateLocation(minimalLocation)
	assert.NoError(t, err)
	assert.NotNil(t, createdLocation)
	assert.NotEmpty(t, createdLocation.Id)
	assert.Equal(t, minimalLocation.Name, createdLocation.Name)
	assert.Equal(t, minimalLocation.Location, createdLocation.Location)
	assert.True(t, createdLocation.Enabled) // Mock sets enabled to true

	// Clean up
	err = client.DeleteLocation(*createdLocation)
	assert.NoError(t, err)
}

func TestLocationLifecycle_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	// Create a test location
	testLocation := Location{
		Name:        "Lifecycle Test Location",
		Description: "Location for lifecycle testing",
		Enabled:     true,
		Location:    "lifecycle-test-region",
	}

	// Create the location
	createdLocation, err := client.CreateLocation(testLocation)
	assert.NoError(t, err)
	assert.NotNil(t, createdLocation)
	assert.NotEmpty(t, createdLocation.Id)

	// Note: The mock server returns a fixed list, so we verify the created location structure
	assert.Equal(t, testLocation.Name, createdLocation.Name)
	assert.Equal(t, testLocation.Description, createdLocation.Description)
	assert.Equal(t, testLocation.Enabled, createdLocation.Enabled)
	assert.Equal(t, testLocation.Location, createdLocation.Location)

	// Delete the location
	err = client.DeleteLocation(*createdLocation)
	assert.NoError(t, err)
}
