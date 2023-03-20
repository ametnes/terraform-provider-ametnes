package ametnes

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const DefaultProductCode = 3795211474

func resourceService() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceCreate,
		ReadContext:   resourceServiceOrNetworkRead,
		DeleteContext: resourceServiceOrNetworkDelete,

		Schema: map[string]*schema.Schema{

			"project": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // if the project changes then we need to force new resource
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // if the name changes the we need to create a new resource
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"kind": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"capacity": {
				Type:     schema.TypeList,
				Required: false,
				ForceNew: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},

						"memory": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"storage": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"nodes": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"config": {
				Type:     schema.TypeMap,
				Required: false,
				ForceNew: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// computed
			"network": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connections": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	client := m.(*Client)

	projectID, err := strconv.Atoi(d.Get("project").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	description := ""

	if desc, ok := d.GetOk("description"); ok {
		description = desc.(string)
	}
	kind := d.Get("kind").(string)

	// get the capacity
	capacity, err := expandCapacitySchema(d.Get("capacity").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	var config map[string]interface{}
	if v, ok := d.GetOk("config"); ok {
		config = v.(map[string]interface{})
	}

	// we add service as prefix for service resource as thats how
	// server differentiates from other resources like network.
	prefixedKind := fmt.Sprintf("service/%s", kind)
	resource := Resource{
		Project:     projectID,
		Kind:        prefixedKind,
		Location:    d.Get("location").(string),
		Name:        d.Get("name").(string),
		Description: description,
		Spec: Spec{
			Components: map[string]interface{}{
				"cpu":     capacity.Cpu,
				"storage": capacity.Storage,
				"memory":  capacity.Memory,
			},
			Nodes:  d.Get("nodes").(int),
			Config: config,
		},
	}

	if networkIntf, ok := d.GetOk("network"); ok {
		networkStr := networkIntf.(string)
		networkInt, err := strconv.Atoi(networkStr)
		if err != nil {
			return diag.FromErr(err)
		}
		resource.Network = networkInt
		resource.Spec.Networks = []Networks{
			{
				Id: networkInt,
			},
		}
	}
	service, err := client.CreateResource(resource)

	if err != nil {
		return diag.FromErr(err)
	}

	respChan := client.checkStatus(projectID, service.Id)
	select {
	case res := <-respChan:
		if res.Success {
			// Identity function
			d.SetId(fmt.Sprintf("%d/%d", projectID, service.Id))
			return resourceServiceOrNetworkRead(ctx, d, m)
		}
	case <-time.After(45 * time.Minute):
		return diag.Errorf("Timeout occured while checking for state")
	}

	// we will not reach here
	return nil
}

func resourceServiceOrNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	client := m.(*Client)

	ids := strings.Split(d.Id(), "/")

	projectID, err := strconv.Atoi(ids[0])
	if err != nil {
		return diag.FromErr(err)
	}
	resourceID, err := strconv.Atoi(ids[1])

	if err != nil {
		return diag.FromErr(err)
	}

	resource, err := client.GetResource(projectID, resourceID)
	if err != nil {
		// if we get error while getting resource then
		d.SetId("")
		return nil
	}
	d.Set("network", resource.Network)
	d.Set("status", resource.Status)
	d.Set("account", resource.Account)

	connections := []Connection{}
	if resource.Spec.Connections != nil && len(resource.Spec.Connections) != 0 {
		connections = resource.Spec.Connections
	} else if resource.Spec.Connection != nil && resource.Spec.Connection.Host != "" {
		connections = append(connections, resource.Spec.Connection)
	}

	if len(connections) != 0 {
		d.Set("connections", connections)
	}

	return nil
}

func resourceServiceOrNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	ids := strings.Split(d.Id(), "/")

	projectID, err := strconv.Atoi(ids[0])
	if err != nil {
		return diag.FromErr(err)
	}
	resourceID, err := strconv.Atoi(ids[1])

	if err != nil {
		return diag.FromErr(err)
	}
	err = client.DeleteResource(Resource{
		Project: projectID,
		Id:      resourceID,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	respChan := client.checkStatusDelete(projectID, resourceID)
	select {
	case res := <-respChan:
		if res.Success {
			// Identity function
			d.SetId("")
			return nil
		}
	case <-time.After(10 * time.Minute):
		return diag.Errorf("Timeout occured while checking for state")
	}
	// we will not get here
	return nil
}

func expandCapacitySchema(in []interface{}) (*Capacity, error) {
	cap := &Capacity{}
	if len(in) == 0 || in[0] == nil {
		return &Capacity{
			Cpu:     1,
			Memory:  1,
			Storage: 1,
		}, nil
	}
	m := in[0].(map[string]interface{})

	if cpu, ok := m["cpu"]; ok {
		cap.Cpu = cpu.(int)
	} else {
		cap.Cpu = 1
	}

	if mem, ok := m["memory"]; ok {
		cap.Memory = mem.(int)
	} else {
		cap.Memory = 1
	}

	if storage, ok := m["storage"]; ok {
		cap.Storage = storage.(int)
	} else {
		cap.Storage = 1
	}

	return cap, nil
}
