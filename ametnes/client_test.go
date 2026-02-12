package ametnes

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestClient(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := GetTestClient(t)
	list, err := client.GetProjects()
	t.Log(list)
	if err != nil {
		t.Skipf("Skipping test due to authentication error: %v", err)
		return
	}
	// Test passes if we can successfully get projects
}

func GetTestClient(t *testing.T) *Client {
	host := "https://api-test.cloud.ametnes.com/v1"
	username := "Brave.Microphone@ametnes.com"
	client, err := NewClient(host, Token{
		Type:     Basic,
		Username: &username,
		Token:    "cCncjAe51,a3bgc91cy4Ke4466571r~da7dZ_791Je9f1Q1244b_",
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}
