package ametnes

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
# curl -k -X GET "https://api-test.cloud.ametnes.com/v1/metadata/locations"
{
  "count": 2,
  "results": [
    {
      "create_date": "2020-12-19 09:02:15",
      "enabled": true,
      "id": "gcp.europe-west2",
      "name": "London, U.K.",
      "provider": "Google Cloud",
      "region": "gcp/europe-west2",
      "update_date": "2020-12-19 09:02:15"
    },
    {
      "create_date": "2020-12-19 09:02:15",
      "enabled": true,
      "id": "aws.eu-west-2",
      "name": "Europe (London)",
      "provider": "Amazon Web Service",
      "region": "aws/eu-west-2",
      "update_date": "2020-12-19 09:02:15"
    }
  ]
}
*/

func TestKindData(t *testing.T) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resource := dataSourceKinds()
	resourceData := schema.TestResourceDataRaw(t, resource.Schema, nil)
	diag := dataSourceKindsRead(context.TODO(), resourceData, nil)
	if diag.HasError() {
		t.Fail()
	}
}
