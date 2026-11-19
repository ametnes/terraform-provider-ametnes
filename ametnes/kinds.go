package ametnes

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetKinds() (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/metadata/resources/kinds", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	kinds := map[string]interface{}{}
	if err := json.Unmarshal(body, &kinds); err != nil {
		return nil, err
	}

	return kinds, nil
}
