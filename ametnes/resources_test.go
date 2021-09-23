package ametnes

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResources(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := GetTestClient(t)

	projects, err := client.GetProjects()
	assert.Nil(t, err)

	assert.Greater(t, len(projects), 0)
	project := projects[0]

	resources, err := client.GetResources(&project)
	assert.Nil(t, err)
	assert.NotNil(t, resources)
}

func TestCreateAndGetResource(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := GetTestClient(t)

	projects, err := client.GetProjects()
	assert.Nil(t, err)

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
	assert.Nil(t, err)

	if assert.NotNil(t, n_resource) {

		// now we know that object isn't nil, we are safe to make
		// further assertions without causing any errors
		assert.Equal(t, "INIT", n_resource.Status)

	}

	gresource, err2 := client.GetResource(project.Id, n_resource.Id)

	assert.Nil(t, err2)

	assert.Equal(t, gresource.Kind, resource.Kind)
}
