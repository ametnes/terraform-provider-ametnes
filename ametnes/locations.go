package ametnes

import (
	"encoding/json"
	"fmt"
	"net/http"
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
