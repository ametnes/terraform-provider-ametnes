package ametnes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// GetMockClient creates a client that uses a mock HTTP server
func GetMockClient(t *testing.T) (*Client, *httptest.Server) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle different endpoints
		switch {
		case r.Method == "GET" && r.URL.Path == "/projects":
			projects := Projects{
				Count: 2,
				Items: []Project{
					{Id: 1, Name: "Test Project 1", Account: 100, Enabled: true},
					{Id: 2, Name: "Test Project 2", Account: 100, Enabled: true},
				},
			}
			json.NewEncoder(w).Encode(projects)

		case r.Method == "POST" && r.URL.Path == "/projects":
			var project Project
			json.NewDecoder(r.Body).Decode(&project)
			project.Id = 999
			project.Account = 100
			project.Enabled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(project)

		case r.Method == "GET" && r.URL.Path == "/metadata/locations":
			locations := Locations{
				Count: 2,
				Items: []Location{
					{Id: "gcp.europe-west2", Name: "London, U.K.", Location: "gcp/europe-west2", Enabled: true, Status: "ONLINE"},
					{Id: "aws.eu-west-2", Name: "Europe (London)", Location: "aws/eu-west-2", Enabled: true, Status: "ONLINE"},
				},
			}
			json.NewEncoder(w).Encode(locations)

		case r.Method == "POST" && r.URL.Path == "/metadata/locations":
			var location Location
			json.NewDecoder(r.Body).Decode(&location)
			location.Id = "test.location"
			location.Enabled = true
			location.Status = "ONLINE"
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(location)

		case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/metadata/locations/"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

		case r.Method == "GET" && r.URL.Path == "/metadata/resources/kinds":
			kinds := map[string]interface{}{
				"count": 2,
				"results": []map[string]interface{}{
					{
						"id":       "mysql:8.0",
						"name":     "MySQL 8.0",
						"type":     "service/mysql:8.0",
						"enabled":  true,
						"release":  "8.0",
						"kind":     "mysql",
						"locations": []string{"gcp/europe-west2"},
						"limits": map[string]interface{}{
							"nodes": []int{1, 2, 3},
						},
						"backups": map[string]interface{}{},
					},
					{
						"id":       "grafana:9.3",
						"name":     "Grafana 9.3",
						"type":     "service/grafana:9.3",
						"enabled":  true,
						"release":  "9.3",
						"kind":     "grafana",
						"locations": []string{"gcp/europe-west2", "aws/eu-west-2"},
						"limits": map[string]interface{}{
							"nodes": []int{1},
						},
						"backups": map[string]interface{}{},
					},
				},
			}
			json.NewEncoder(w).Encode(kinds)

		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/projects/") && strings.HasSuffix(r.URL.Path, "/resources"):
			// Extract project ID from path like "/projects/1/resources"
			resources := Resources{
				Count: 1,
				Items: []Resource{
					{
						Id:          1,
						Project:     1,
						Account:     100,
						Kind:        "service/mysql:8.0",
						Location:    "gcp/europe-west2",
						Name:        "Test Resource",
						Status:      "ONLINE",
						Description: "Test Description",
						Network:     1,
						Spec: Spec{
							Components: map[string]interface{}{
								"storage": 1,
							},
							Nodes: 1,
							Config: map[string]interface{}{
								"test.config": "value",
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resources)

		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/projects/") && strings.Contains(r.URL.Path, "/resources/") && !strings.HasSuffix(r.URL.Path, "/resources"):
			// Extract project ID and resource ID from path like "/projects/1/resources/1"
			resource := Resource{
				Id:          1,
				Project:     1,
				Account:     100,
				Kind:        "service/mysql:8.0",
				Location:    "gcp/europe-west2",
				Name:        "Test Resource",
				Status:      "ONLINE",
				Description: "Test Description",
				Network:     1,
				Spec: Spec{
					Components: map[string]interface{}{
						"storage": 1,
					},
					Nodes: 1,
					Config: map[string]interface{}{
						"test.config": "value",
					},
					Connection: Connection{
						Host: "test.example.com",
						Port: 3306,
						Name: "mysql",
					},
				},
			}
			json.NewEncoder(w).Encode(resource)

		case r.Method == "POST" && strings.HasPrefix(r.URL.Path, "/projects/") && strings.HasSuffix(r.URL.Path, "/resources"):
			var resource Resource
			json.NewDecoder(r.Body).Decode(&resource)
			resource.Id = 999
			resource.Status = "INIT"
			resource.Account = 100
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resource)

		case r.Method == "PUT" && strings.HasPrefix(r.URL.Path, "/projects/"):
			var resource Resource
			json.NewDecoder(r.Body).Decode(&resource)
			resource.Status = "ONLINE"
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resource)

		case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/projects/") && strings.Contains(r.URL.Path, "/resources/"):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("Mock server: endpoint not found: %s %s", r.Method, r.URL.Path),
			})
		}
	}))

	// Create client pointing to mock server
	username := "test-user"
	client, err := NewClient(server.URL, Token{
		Type:     Basic,
		Username: &username,
		Token:    "test-token",
	})
	if err != nil {
		t.Fatal(err)
	}

	return client, server
}

