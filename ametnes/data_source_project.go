package ametnes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceProject() *schema.Resource {
	return &schema.Resource{
		Description: `
Read an Ametnes cloud project resource.
`,
		ReadContext: dataSourceProjectRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Name of the project.",
			},
			"account_id": {
				Type:     schema.TypeInt,
				Computed: true,
				Description: "Ametnes cloud account id of the project.",
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "`false` if this project is disabled.",
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Description of the project.",
			},
		},
	}
}

func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	projects, err := client.GetProjects()
	if err != nil {
		// Check if this is an authentication error
		errMsg := err.Error()
		if strings.Contains(errMsg, "401") || 
			strings.Contains(errMsg, "Unauthorized") ||
			strings.Contains(errMsg, "status: 401") {
			return diag.Errorf("Authentication failed: %v", err)
		}
		return diag.FromErr(err)
	}

	projectID := -1
	var dProject Project
	projectName := d.Get("name").(string)

	for _, project := range projects {
		if project.Name == projectName {
			projectID = project.Id
			dProject = project
			break
		}
	}
	if projectID == -1 {
		return diag.Errorf("Cannot find your project with name %s", projectName)
	}

	d.SetId(fmt.Sprint(projectID))
	d.Set("account_id", dProject.Account)
	d.Set("enabled", dProject.Enabled)
	d.Set("description", dProject.Description)

	return nil

}
