package ametnes

import (
	"crypto/tls"
	"net/http"
	"testing"
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
	t.Log(err)

}

func GetTestClient(t *testing.T) *Client {
	host := "https://cloud.ametnes.com/api"
	client, err := NewClient(host, Token{
		Type:  Bearer,
		Token: "YR7gASZaRCgJ3Evf78N5kz3oCDUwlJYoaunj7EEc8HV8S5ypm9hmhaH1IQRyMM1K1L6XSJKbuxevfyriSwvDPMxgZXUbWqgbxofKL9XgpSV800ou5cI9juwHivVwSPAe",
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}
