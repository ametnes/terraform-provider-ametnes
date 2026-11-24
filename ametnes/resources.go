package ametnes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) GetResources(project *Project) ([]Resource, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%d/resources", c.HostURL, project.Id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	resources := Resources{}
	err = json.Unmarshal(body, &resources)
	if err != nil {
		return nil, err
	}

	return resources.Items, nil
}

func (c *Client) GetResource(projectId, id int) (*Resource, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%d/resources/%d", c.HostURL, projectId, id), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	resource := Resource{}
	err = json.Unmarshal(body, &resource)
	if err != nil {
		return nil, err
	}

	return &resource, nil
}

// GetResourceFromProject finds a resource in its project's resource list.
//
// It reports the resource's current lifecycle status and returns nil once the
// resource is no longer present. Use it when the lifecycle status matters,
// such as waiting for a delete to complete.
func (c *Client) GetResourceFromProject(projectID, resourceID int) (*Resource, error) {
	resources, err := c.GetResources(&Project{Id: projectID})
	if err != nil {
		return nil, err
	}

	for i := range resources {
		if resources[i].Id == resourceID {
			return &resources[i], nil
		}
	}

	return nil, nil
}

func (c *Client) CreateResource(resource Resource) (*Resource, error) {
	rb, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects/%d/resources", c.HostURL, resource.Project), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	n_resource := Resource{}
	err = json.Unmarshal(body, &n_resource)
	if err != nil {
		return nil, err
	}

	return &n_resource, nil
}

func (c *Client) UpdateResource(resource Resource) (*Resource, error) {
	rb, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/projects/%d/resources/%d", c.HostURL, resource.Project, resource.Id), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	n_resource := Resource{}
	err = json.Unmarshal(body, &n_resource)
	if err != nil {
		return nil, err
	}

	return &n_resource, nil
}

func (c *Client) DeleteResource(resource Resource, deleteBackups bool) error {
	url := fmt.Sprintf("%s/projects/%d/resources/%d", c.HostURL, resource.Project, resource.Id)
	if deleteBackups {
		// Without this flag the service's backups are retained for the
		// location's backup retention window; the resource stays DELETED
		// until that window expires.
		url = fmt.Sprintf("%s?delete_backups=true", url)
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return err
	}

	n_resource := Resource{}
	err = json.Unmarshal(body, &n_resource)
	if err != nil {
		return err
	}

	return nil
}
