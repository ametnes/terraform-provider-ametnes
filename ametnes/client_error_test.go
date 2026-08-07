package ametnes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func getMockClientForStatus(status string) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "GET" && r.URL.Path == "/projects/1/resources/1":
			resource := Resource{
				Id:      1,
				Project: 1,
				Account: 100,
				Status:  status,
			}
			json.NewEncoder(w).Encode(resource)

		case r.Method == "POST" && r.URL.Path == "/projects/1/resources":
			var resource Resource
			json.NewDecoder(r.Body).Decode(&resource)
			resource.Id = 999
			resource.Status = "INIT"
			resource.Account = 100
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resource)

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "not found",
			})
		}
	}))

	username := "test-user"
	client, _ := NewClient(server.URL, Token{
		Type:     Basic,
		Username: &username,
		Token:    "test-token",
	})

	return client, server
}

func TestCheckStatus_ReturnsErrorOnErrorState(t *testing.T) {
	client, server := getMockClientForStatus("ERROR")
	defer server.Close()

	respChan := client.checkStatus(1, 1)

	select {
	case res := <-respChan:
		assert.False(t, res.Success, "expected Success to be false on ERROR status")
		assert.Error(t, res.Error, "expected an error on ERROR status")
		assert.Contains(t, res.Error.Error(), "ERROR state")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for checkStatus to return")
	}
}

func TestCheckStatus_ReturnsSuccessOnActiveState(t *testing.T) {
	client, server := getMockClientForStatus("ACTIVE")
	defer server.Close()

	respChan := client.checkStatus(1, 1)

	select {
	case res := <-respChan:
		assert.True(t, res.Success, "expected Success to be true on ACTIVE status")
		assert.NoError(t, res.Error, "expected no error on ACTIVE status")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for checkStatus to return")
	}
}

func TestCheckStatus_ContinuesPollingOnInitState(t *testing.T) {
	client, server := getMockClientForStatus("INIT")
	defer server.Close()

	respChan := client.checkStatus(1, 1)

	select {
	case <-respChan:
		t.Fatal("checkStatus should not return immediately for INIT status")
	case <-time.After(2 * time.Second):
	}
}

func TestCheckStatus_ReturnsErrorOnGetResourceFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	username := "test-user"
	client, _ := NewClient(server.URL, Token{
		Type:     Basic,
		Username: &username,
		Token:    "test-token",
	})

	respChan := client.checkStatus(1, 1)
	defer server.Close()

	select {
	case res := <-respChan:
		assert.False(t, res.Success, "expected Success to be false on server error")
		assert.Error(t, res.Error, "expected an error on server failure")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for checkStatus to return")
	}
}

func TestCheckStatusDelete_ReturnsErrorOnErrorState(t *testing.T) {
	client, server := getMockClientForStatus("ERROR")
	defer server.Close()

	respChan := client.checkStatusDelete(1, 1)

	select {
	case res := <-respChan:
		assert.False(t, res.Success, "expected Success to be false on ERROR status")
		assert.Error(t, res.Error, "expected an error on ERROR status during delete")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for checkStatusDelete to return")
	}
}

func TestResourceServiceCreate_FailsOnErrorState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/projects/1/resources":
			resource := Resource{
				Id:      999,
				Project: 1,
				Account: 100,
				Status:  "INIT",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resource)

		case r.Method == "GET" && r.URL.Path == "/projects/1/resources/999":
			resource := Resource{
				Id:      999,
				Project: 1,
				Account: 100,
				Status:  "ERROR",
			}
			json.NewEncoder(w).Encode(resource)

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
		}
	}))
	defer server.Close()

	username := "test-user"
	client, err := NewClient(server.URL, Token{
		Type:     Basic,
		Username: &username,
		Token:    "test-token",
	})
	assert.NoError(t, err)

	rsrc := resourceService()
	d := schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"project":     "1",
		"name":        "test-service",
		"kind":        "mysql:8.0",
		"location":    "gcp/europe-west2",
		"network":     "1",
		"description": "test",
		"nodes":       1,
		"config":      map[string]interface{}{},
		"capacity":    []interface{}{},
	})

	diags := resourceServiceCreate(context.Background(), d, client)
	assert.True(t, diags.HasError(), "expected diagnostics to have error when resource enters ERROR state")
	assert.Contains(t, diags[0].Summary, "ERROR state")
}

func TestResourceNetworkCreate_FailsOnErrorState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/projects/1/resources":
			resource := Resource{
				Id:      998,
				Project: 1,
				Account: 100,
				Status:  "INIT",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resource)

		case r.Method == "GET" && r.URL.Path == "/projects/1/resources/998":
			resource := Resource{
				Id:      998,
				Project: 1,
				Account: 100,
				Status:  "ERROR",
			}
			json.NewEncoder(w).Encode(resource)

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
		}
	}))
	defer server.Close()

	username := "test-user"
	client, err := NewClient(server.URL, Token{
		Type:     Basic,
		Username: &username,
		Token:    "test-token",
	})
	assert.NoError(t, err)

	rsrc := resourceNetwork()
	d := schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"project":     "1",
		"name":        "test-network",
		"description": "test",
		"location":    "gcp/europe-west2",
	})

	diags := resourceNetworkCreate(context.Background(), d, client)
	assert.True(t, diags.HasError(), "expected diagnostics to have error when network enters ERROR state")
	assert.Contains(t, diags[0].Summary, "ERROR state")
}
