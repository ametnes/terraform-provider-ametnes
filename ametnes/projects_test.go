package ametnes

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestProjects(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	// host := "https://api-test.cloud.ametnes.com/v1"
	// username := "Brave.Microphone@ametnes.com"
	// token := "a03of\\75Ven4ada7A0W1h1>21f=4}b5fadQdn458254e@b3Tb\\"

	client, err := NewClient(&Host, &UserName, &Token)

	if err != nil {
		t.Fail()
	}

	projects, err := client.GetProjects()
	if err != nil {
		t.Fail()
	}
	if len(projects) == 0 {
		t.Fail()
	}
}
