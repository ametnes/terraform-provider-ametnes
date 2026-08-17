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

func TestCheckStatus_ReturnsSuccessOnReadyState(t *testing.T) {
	client, server := getMockClientForStatus("READY")
	defer server.Close()

	respChan := client.checkStatus(1, 1)

	select {
	case res := <-respChan:
		assert.True(t, res.Success, "expected Success to be true on READY status")
		assert.NoError(t, res.Error, "expected no error on READY status")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for checkStatus to return")
	}
}

func TestCheckStatus_ReturnsSuccessOnInitializedState(t *testing.T) {
	client, server := getMockClientForStatus("INITIALIZED")
	defer server.Close()

	respChan := client.checkStatus(1, 1)

	select {
	case res := <-respChan:
		assert.True(t, res.Success, "expected Success to be true on INITIALIZED status")
		assert.NoError(t, res.Error, "expected no error on INITIALIZED status")
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

func TestCheckStatus_ContinuesPollingOnUpdatingState(t *testing.T) {
	client, server := getMockClientForStatus("UPDATING")
	defer server.Close()

	respChan := client.checkStatus(1, 1)

	select {
	case <-respChan:
		t.Fatal("checkStatus should not return immediately for UPDATING status")
	case <-time.After(2 * time.Second):
	}
}

func TestCheckStatus_ReturnsSuccessAfterUpdatingTransitions(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		status := "UPDATING"
		if callCount > 1 {
			status = "READY"
		}
		resource := Resource{
			Id:      1,
			Project: 1,
			Account: 100,
			Status:  status,
		}
		json.NewEncoder(w).Encode(resource)
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
		assert.True(t, res.Success, "expected Success after UPDATING transitions to READY")
		assert.NoError(t, res.Error)
		assert.GreaterOrEqual(t, callCount, 2, "expected multiple polls before status transitioned")
	case <-time.After(35 * time.Second):
		t.Fatal("timed out waiting for status transition from UPDATING to READY")
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

func getMockClientForUpdate() (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "GET" && r.URL.Path == "/projects/1/resources/999":
			resource := Resource{
				Id:          999,
				Project:     1,
				Account:     100,
				Kind:        "service/mysql:8.0",
				Location:    "gcp/europe-west2",
				Name:        "test-service",
				Status:      "READY",
				Network:     1,
				Spec: Spec{
					Components: map[string]interface{}{
						"cpu":     1,
						"memory":  1,
						"storage": 1,
					},
					Nodes:   1,
					Config:  map[string]interface{}{},
					Networks: []Networks{{Id: 1}},
					Connection: Connection{
						Host: "test.example.com",
						Port: 3306,
						Name: "mysql",
					},
				},
			}
			json.NewEncoder(w).Encode(resource)

		case r.Method == "PUT" && r.URL.Path == "/projects/1/resources/999":
			var resource Resource
			json.NewDecoder(r.Body).Decode(&resource)
			resource.Status = "READY"
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resource)

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "not found"})
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

func TestResourceServiceUpdate_WaitsForStatusAfterUpdate(t *testing.T) {
	client, server := getMockClientForUpdate()
	defer server.Close()

	rsrc := resourceService()
	d := schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"project":     "1",
		"name":        "test-service",
		"kind":        "mysql:8.0",
		"location":    "gcp/europe-west2",
		"description": "test",
		"nodes":       1,
		"config":      map[string]interface{}{"key": "old-value"},
		"capacity":    []interface{}{},
		"alias":       "test-alias",
	})
	d.SetId("1/999")
	d.Set("config", map[string]interface{}{"key": "new-value"})

	diags := resourceServiceUpdate(context.Background(), d, client)
	assert.False(t, diags.HasError(), "expected update to succeed and poll for status")
	assert.Equal(t, "1/999", d.Id())
}

func TestResourceNetworkUpdate_WaitsForStatusAfterUpdate(t *testing.T) {
	client, server := getMockClientForUpdate()
	defer server.Close()

	rsrc := resourceNetwork()
	d := schema.TestResourceDataRaw(t, rsrc.Schema, map[string]interface{}{
		"project":     "1",
		"name":        "test-network",
		"description": "test",
		"location":    "gcp/europe-west2",
		"kind":        "loadbalancer:1.0",
	})
	d.SetId("1/999")
	d.Set("name", "updated-network-name")

	diags := resourceNetworkUpdate(context.Background(), d, client)
	assert.False(t, diags.HasError(), "expected network update to succeed and poll for status")
	assert.Equal(t, "1/999", d.Id())
}

func TestResourceServiceCreate_DefaultsPublicVisibleToTrue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/projects/1/resources":
			var resource Resource
			json.NewDecoder(r.Body).Decode(&resource)
			publicVisible, ok := resource.Spec.Config["public.visible"]
			assert.True(t, ok, "expected public.visible to be present in config")
			assert.Equal(t, "true", publicVisible, "expected public.visible to default to 'true'")
			resource.Id = 999
			resource.Status = "READY"
			resource.Account = 100
			resource.Spec.Connection = Connection{
				Host: "test.example.com",
				Port: 3306,
				Name: "mysql",
			}
			resource.Spec.Networks = []Networks{{Id: 1}}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resource)

		case r.Method == "GET" && r.URL.Path == "/projects/1/resources/999":
			resource := Resource{
				Id:          999,
				Project:     1,
				Account:     100,
				Kind:        "service/mysql:8.0",
				Location:    "gcp/europe-west2",
				Name:        "test-service",
				Status:      "READY",
				Network:     1,
				Spec: Spec{
					Components: map[string]interface{}{
						"cpu":     1,
						"memory":  1,
						"storage": 1,
					},
					Nodes:   1,
					Config:  map[string]interface{}{"public.visible": "true"},
					Networks: []Networks{{Id: 1}},
					Connection: Connection{
						Host: "test.example.com",
						Port: 3306,
						Name: "mysql",
					},
				},
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
		"capacity":    []interface{}{},
	})

	diags := resourceServiceCreate(context.Background(), d, client)
	assert.False(t, diags.HasError(), "expected create to succeed with default public.visible")
}

func TestResourceServiceCreate_PreservesExplicitPublicVisible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/projects/1/resources":
			var resource Resource
			json.NewDecoder(r.Body).Decode(&resource)
			publicVisible := resource.Spec.Config["public.visible"]
			assert.Equal(t, "false", publicVisible, "expected explicit public.visible to be preserved")
			resource.Id = 999
			resource.Status = "READY"
			resource.Account = 100
			resource.Spec.Connection = Connection{
				Host: "test.example.com",
				Port: 3306,
				Name: "mysql",
			}
			resource.Spec.Networks = []Networks{{Id: 1}}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resource)

		case r.Method == "GET" && r.URL.Path == "/projects/1/resources/999":
			resource := Resource{
				Id:          999,
				Project:     1,
				Account:     100,
				Kind:        "service/mysql:8.0",
				Location:    "gcp/europe-west2",
				Name:        "test-service",
				Status:      "READY",
				Network:     1,
				Spec: Spec{
					Components: map[string]interface{}{
						"cpu":     1,
						"memory":  1,
						"storage": 1,
					},
					Nodes:   1,
					Config:  map[string]interface{}{"public.visible": "false"},
					Networks: []Networks{{Id: 1}},
					Connection: Connection{
						Host: "test.example.com",
						Port: 3306,
						Name: "mysql",
					},
				},
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
		"config":      map[string]interface{}{"public.visible": "false"},
		"capacity":    []interface{}{},
	})

	diags := resourceServiceCreate(context.Background(), d, client)
	assert.False(t, diags.HasError(), "expected create to succeed with explicit public.visible")
}

func TestResourceServiceCreate_DefaultsPublicVisibleWithNilConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/projects/1/resources":
			var resource Resource
			json.NewDecoder(r.Body).Decode(&resource)
			publicVisible, ok := resource.Spec.Config["public.visible"]
			assert.True(t, ok, "expected public.visible to be set when config is nil")
			assert.Equal(t, "true", publicVisible, "expected public.visible to default to 'true'")
			resource.Id = 999
			resource.Status = "READY"
			resource.Account = 100
			resource.Spec.Connection = Connection{
				Host: "test.example.com",
				Port: 3306,
				Name: "mysql",
			}
			resource.Spec.Networks = []Networks{{Id: 1}}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resource)

		case r.Method == "GET" && r.URL.Path == "/projects/1/resources/999":
			resource := Resource{
				Id:          999,
				Project:     1,
				Account:     100,
				Kind:        "service/mysql:8.0",
				Location:    "gcp/europe-west2",
				Name:        "test-service",
				Status:      "READY",
				Network:     1,
				Spec: Spec{
					Components: map[string]interface{}{
						"cpu":     1,
						"memory":  1,
						"storage": 1,
					},
					Nodes:   1,
					Config:  map[string]interface{}{"public.visible": "true"},
					Networks: []Networks{{Id: 1}},
					Connection: Connection{
						Host: "test.example.com",
						Port: 3306,
						Name: "mysql",
					},
				},
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
		"capacity":    []interface{}{},
	})

	diags := resourceServiceCreate(context.Background(), d, client)
	assert.False(t, diags.HasError(), "expected create with nil config to succeed with public.visible default")
}
