package ametnes

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// APIError is returned when the Ametnes API responds with a non-200 status.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status: %d, body: %s", e.StatusCode, e.Body)
}

func isNotFoundError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// isResourceGone reports whether the error means the resource no longer exists.
// A deleted resource is reported as 404. Older cloud API builds surface the same
// condition as a 400 "not found" message, so accept that too.
func isResourceGone(err error) bool {
	if isNotFoundError(err) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusBadRequest {
		body := strings.ToLower(apiErr.Body)
		return strings.Contains(body, "resource") && strings.Contains(body, "not found")
	}
	return false
}

// isPermanentAPIError reports whether retrying the request cannot succeed, such
// as an authentication, authorization or malformed request failure. Server
// errors and transport failures are treated as transient.
func isPermanentAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500
}

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
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
		HostURL:    host,
	}

	c.AuthHeader = tkn.GetAuthHeader()
	return &c, nil
}

func (c *Client) SetHTTPTimeout(timeout int) {
	c.HTTPClient.Timeout = time.Second * time.Duration(timeout)
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
		return nil, &APIError{StatusCode: res.StatusCode, Body: string(body)}
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
			if resource.Status == "ERROR" {
				respChan <- Status{
					Success: false,
					Error:   fmt.Errorf("resource %d entered ERROR state", resourceID),
				}
				close(respChan)
				return
			}
			if resource.Status == "READY" || resource.Status == "INITIALIZED" {
				respChan <- Status{
					Success: true,
				}
				close(respChan)
				return
			}
			time.Sleep(30 * time.Second)
		}

	}()

	return respChan
}

func (c *Client) checkStatusDelete(ctx context.Context, projectID, resourceID int) chan Status {
	respChan := make(chan Status, 1)

	go func() {
		defer close(respChan)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			resource, err := c.GetResource(projectID, resourceID)
			switch {
			case err == nil && resource.Status == "DELETED":
				// The control plane may keep deleted resources as tombstones,
				// so a successful read can still mean the delete is complete.
				respChan <- Status{Success: true}
				return
			case err == nil && resource.Status == "ERROR":
				respChan <- Status{
					Success: false,
					Error:   fmt.Errorf("error while deleting resource id %d", resourceID),
				}
				return
			case isResourceGone(err):
				// The control plane reports the resource as absent once it has
				// been deleted.
				respChan <- Status{Success: true}
				return
			case isPermanentAPIError(err):
				// Authentication/authorization or a malformed request will not
				// succeed on retry. Do not report a successful delete while the
				// resource may still exist.
				respChan <- Status{
					Success: false,
					Error:   fmt.Errorf("error while deleting resource id %d: %w", resourceID, err),
				}
				return
			}

			select {
			case <-ctx.Done():
				respChan <- Status{Success: false, Error: ctx.Err()}
				return
			case <-ticker.C:
			}
		}
	}()

	return respChan
}
