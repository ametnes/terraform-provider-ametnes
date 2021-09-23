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
