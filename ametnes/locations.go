package ametnes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) GetLocations() ([]Location, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/metadata/locations", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	data, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	resp := Locations{}
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}

func (c *Client) CreateLocation(location Location) (*Location, error) {
	rb, err := json.Marshal(location)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/metadata/locations", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	data, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	newLoc := Location{}
	err = json.Unmarshal(data, &newLoc)
	if err != nil {
		return nil, err
	}

	return &newLoc, nil
}

func (c *Client) DeleteLocation(location Location) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/metadata/locations/%s", c.HostURL, location.Id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
