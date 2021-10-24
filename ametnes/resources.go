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

func (c *Client) DeleteResource(resource Resource) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/projects/%d/resources/%d", c.HostURL, resource.Project, resource.Id), nil)
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
