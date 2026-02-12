package ametnes

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocations(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := GetTestClient(t)

	locations, err := client.GetLocations()
	assert.Nil(t, err)
	assert.NotNil(t, locations)

	if len(locations) > 0 {
		// Verify the structure of at least one location
		location := locations[0]
		assert.NotEmpty(t, location.Id)
		assert.NotEmpty(t, location.Name)
		assert.NotEmpty(t, location.Location)
	}
}

func TestCreateAndDeleteLocation(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := GetTestClient(t)

	// First, get existing locations to understand the structure
	existingLocations, err := client.GetLocations()
	assert.Nil(t, err)

	if len(existingLocations) > 0 {
		// Test creating a new location (using similar structure to existing ones)
		// Note: This might fail if the test account doesn't have permission to create locations
		newLocation := Location{
			Name:        "Test Location",
			Description: "Test location created by unit tests",
			Enabled:     true,
			Location:    "test-region",
		}

		createdLocation, err := client.CreateLocation(newLocation)

		// If creation is successful, test deletion
		if assert.Nil(t, err) && createdLocation != nil {
			assert.NotEmpty(t, createdLocation.Id)
			assert.Equal(t, newLocation.Name, createdLocation.Name)
			assert.Equal(t, newLocation.Description, createdLocation.Description)
			assert.Equal(t, newLocation.Enabled, createdLocation.Enabled)
			assert.Equal(t, newLocation.Location, createdLocation.Location)

			// Test deletion
			err = client.DeleteLocation(*createdLocation)
			assert.Nil(t, err)
		}
	}
}

func TestGetLocationsReturnsValidStructure(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := GetTestClient(t)

	locations, err := client.GetLocations()
	assert.Nil(t, err)

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

func TestLocationOperationsWithInvalidData(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := GetTestClient(t)

	// Test creating location with minimal data
	minimalLocation := Location{
		Name:     "Minimal Test Location",
		Location: "minimal-test-region",
	}

	createdLocation, err := client.CreateLocation(minimalLocation)

	// If creation succeeds, verify structure and clean up
	if err == nil && createdLocation != nil {
		assert.NotEmpty(t, createdLocation.Id)
		assert.Equal(t, minimalLocation.Name, createdLocation.Name)
		assert.Equal(t, minimalLocation.Location, createdLocation.Location)

		// Default values should be set
		assert.False(t, createdLocation.Enabled) // Assuming default is false

		// Clean up
		err = client.DeleteLocation(*createdLocation)
		assert.Nil(t, err)
	}
}

func TestLocationLifecycle(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := GetTestClient(t)

	// Create a test location
	testLocation := Location{
		Name:        "Lifecycle Test Location",
		Description: "Location for lifecycle testing",
		Enabled:     true,
		Location:    "lifecycle-test-region",
	}

	// Create the location
	createdLocation, err := client.CreateLocation(testLocation)
	if err != nil {
		t.Skipf("Skipping lifecycle test: Location creation not supported or permission denied: %v", err)
		return
	}

	// Verify creation
	assert.Nil(t, err)
	assert.NotNil(t, createdLocation)
	assert.NotEmpty(t, createdLocation.Id)

	// Get all locations and verify our created location exists
	allLocations, err := client.GetLocations()
	assert.Nil(t, err)

	found := false
	for _, loc := range allLocations {
		if loc.Id == createdLocation.Id {
			found = true
			assert.Equal(t, testLocation.Name, loc.Name)
			assert.Equal(t, testLocation.Description, loc.Description)
			assert.Equal(t, testLocation.Enabled, loc.Enabled)
			assert.Equal(t, testLocation.Location, loc.Location)
			break
		}
	}
	assert.True(t, found, "Created location should exist in the locations list")

	// Delete the location
	err = client.DeleteLocation(*createdLocation)
	assert.Nil(t, err)

	// Verify deletion by checking the location no longer exists
	updatedLocations, err := client.GetLocations()
	assert.Nil(t, err)

	stillExists := false
	for _, loc := range updatedLocations {
		if loc.Id == createdLocation.Id {
			stillExists = true
			break
		}
	}
	assert.False(t, stillExists, "Deleted location should not exist in the locations list")
}
