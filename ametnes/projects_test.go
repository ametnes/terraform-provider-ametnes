package ametnes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjects_Unit(t *testing.T) {
	client, server := GetMockClient(t)
	defer server.Close()

	projects, err := client.GetProjects()
	assert.NoError(t, err)
	assert.Greater(t, len(projects), 0)
	assert.Equal(t, "Test Project 1", projects[0].Name)
}
