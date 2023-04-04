package ametnes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Description: `
Creates and manages an ametnes cloud project. All Ametnes cloud resources must be created in a project.
`,
		CreateContext: resourceProjectCreate,
		ReadContext:   dataSourceProjectRead,
		DeleteContext: resourceProjectDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "The name of the project.",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The description of the project.",
			},
			"account_id": {
				Type:     schema.TypeInt,
				Computed: true,
				Description: "The account where the project is created.",
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	name := d.Get("name").(string)
	var description string

	if desc, ok := d.GetOk("description"); ok {
		description = desc.(string)
	}

	project := Project{
		Name:        name,
		Description: description,
	}

	resp, err := client.CreateProject(project)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprint(resp.Id))
	d.Set("account_id", resp.Account)
	d.Set("enabled", resp.Enabled)
	return nil
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	projectID := d.Id()
	iProjectID, err := strconv.Atoi(projectID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.DeleteProject(Project{
		Id: iProjectID,
	})

	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
