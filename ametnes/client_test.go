package ametnes

import (
	"crypto/tls"
	"net/http"
	"testing"
)

var Host = "https://api-test.cloud.ametnes.com/v1"
var UserName = "Brave.Microphone@ametnes.com"
var Token = "1z8a41c24dc2e341TG9qPn5r2k;_6Q6A56G31j970d6b}wq7t@fDco"

func TestClient(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// resource := dataSourceKinds()
	// resourceData := schema.TestResourceDataRaw(t, resource.Schema, nil)
	// diag := dataSourceKindsRead(context.TODO(), resourceData, nil)
	// fmt.Printf("%+v\n", diag)

	// host := "https://api-test.cloud.ametnes.com/v1"
	// username := "Brave.Microphone@ametnes"
	// token := "2au\\a0d1ecm584719Ufmd8b999eYRNxV04e2478!jHUi947H37mC3"

	// // client, err := NewClient(&host, &username, &token)

	// // if err != nil {
	// // 	t.Fail()
	// // }
}
