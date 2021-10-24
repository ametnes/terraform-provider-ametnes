package ametnes

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// resource := dataSourceKinds()
	// resourceData := schema.TestResourceDataRaw(t, resource.Schema, nil)
	// diag := dataSourceKindsRead(context.TODO(), resourceData, nil)
	// fmt.Printf("%+v\n", diag)

	// host := "https://cloud.ametnes.com/api"
	// username := "Brave.Microphone@ametnes.com"
	// token := "2eh,.Uc983QeAfeb<1cT3oeum34SaD7u0(b2dc&-E53acc2ic62y1"

	client := GetTestClient(t)
	list, err := client.GetProjects()
	t.Log(list)
	assert.Nil(t, err)

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
