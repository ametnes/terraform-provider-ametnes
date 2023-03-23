package ametnes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) GetProjects() ([]Project, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	projects := Projects{}
	err = json.Unmarshal(body, &projects)
	if err != nil {
		return nil, err
	}

	return projects.Items, nil
}

func (c *Client) CreateProject(project Project) (*Project, error) {
	rb, err := json.Marshal(project)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/projects", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	newProject := Project{}
	err = json.Unmarshal(body, &newProject)
	if err != nil {
		return nil, err
	}

	return &newProject, nil
}

func (c *Client) UpdateProject(project Project) (*Project, error) {
	rb, err := json.Marshal(project)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/projects/%d", c.HostURL, project.Id), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	n_project := Project{}
	err = json.Unmarshal(body, &n_project)
	if err != nil {
		return nil, err
	}

	return &n_project, nil
}

func (c *Client) DeleteProject(project Project) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/projects/%d", c.HostURL, project.Id), nil)
	if err != nil {
		return err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return err
	}

	n_project := Project{}
	err = json.Unmarshal(body, &n_project)
	if err != nil {
		return err
	}

	return nil
}
