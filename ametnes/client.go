package ametnes

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// HostURL - Default Hashicups URL
// const HostURL string = "http://localhost:19090"

// Client -
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	AuthHeader string
}

type TokenType int

const (
	Basic  TokenType = 0
	Bearer           = iota
)

type Status struct {
	Error   error
	Success bool
}

type Token struct {
	Type TokenType

	Username *string // this would be nil if token type is bearer
	Token    string
}

func (tk *Token) GetAuthHeader() string {
	if tk.Type == Basic {
		UserNameToken := fmt.Sprintf("%s:%s", *tk.Username, tk.Token)
		endcodedToken := base64.StdEncoding.EncodeToString([]byte(UserNameToken))
		return fmt.Sprintf("Basic %s", endcodedToken)
	} else if tk.Type == Bearer {
		return fmt.Sprintf("Bearer %s", tk.Token)
	}
	return ""
}

// NewClient
func NewClient(host string, tkn Token) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    host,
	}

	c.AuthHeader = tkn.GetAuthHeader()
	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", c.AuthHeader)
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

func (c *Client) checkStatus(projectID, resourceID int) chan Status {
	respChan := make(chan Status)

	go func() {
		for {
			resource, err := c.GetResource(projectID, resourceID)
			if err != nil {
				respChan <- Status{
					Error:   err,
					Success: false,
				}
				close(respChan)
				return
			}
			if resource.Status == "READY" {
				respChan <- Status{
					Success: true,
				}
				close(respChan)
				return
			}

			if resource.Status == "ERROR" {
				respChan <- Status{
					Success: false,
					Error:   fmt.Errorf("error while creating resource %d", resourceID),
				}
			}

			time.Sleep(30 * time.Second)
		}

	}()

	return respChan
}

func (c *Client) checkStatusDelete(projectID, resourceID int) chan Status {
	respChan := make(chan Status)

	go func() {
		for {
			resource, err := c.GetResource(projectID, resourceID)
			// if there is an error while getting the resource then mostly the resource is deleted
			if err != nil {
				respChan <- Status{
					Success: true,
				}
				close(respChan)
				return
			}

			if resource != nil && resource.Status == "ERROR" {
				respChan <- Status{
					Success: false,
					Error:   fmt.Errorf("error while deleting resource id %d", resourceID),
				}
				close(respChan)
				return
			}

			time.Sleep(10 * time.Second)
		}

	}()

	return respChan
}
