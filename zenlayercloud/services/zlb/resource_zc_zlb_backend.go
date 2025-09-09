package zlb

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zlb "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zlb20250401"
	"time"
)

func ResourceZenlayerCloudZlbBackend() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZlbBackendCreate,
		ReadContext:   resourceZenlayerCloudZlbBackendRead,
		UpdateContext: resourceZenlayerCloudZlbBackendUpdate,
		DeleteContext: resourceZenlayerCloudZlbBackendDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"zlb_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the load balancer instance.",
			},
			"listener_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the listener.",
			},
			"backends": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of backend servers.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "ID of the backend server. The added instance must belong to the VPC associated with lb.",
						},
						"private_ip_address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Private IP address of the network interface attached to the instance.",
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "Target port for request forwarding and health checks. If left empty, it will follow the listener's port configuration. Valid values: `1` to `65535`.",
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      100,
							Description:  "Forwarding weight of the backend server. Valid value ranges: (0~65535). Default to 100. Weight of 0 means the server will not accept new requests.",
							ValidateFunc: validation.IntBetween(0, 65535),
						},
					},
				},
			},
		},
	}
}

func resourceZenlayerCloudZlbBackendCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb_backend.create")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zlb.NewRegisterBackendRequest()
	request.LoadBalancerId = common.String(d.Get("zlb_id").(string))
	request.ListenerId = common.String(d.Get("listener_id").(string))

	// Process backends
	backends := d.Get("backends").(*schema.Set).List()
	backendList := make([]*zlb.BackendServer, 0, len(backends))

	for _, backend := range backends {
		item := backend.(map[string]interface{})
		backendItem := &zlb.BackendServer{
			InstanceId:       common.String(item["instance_id"].(string)),
			PrivateIpAddress: common.String(item["private_ip_address"].(string)),
		}

		if weight, exists := item["weight"]; exists {
			v := weight.(int)
			backendItem.Weight = common.Integer(v)
		}

		if port, exists := item["port"]; exists {
			v := port.(int)
			if v != 0 {
				backendItem.Port = common.Integer(v)
			}
		}

		backendList = append(backendList, backendItem)
	}

	request.BackendServers = backendList

	_, err := zlbService.client.WithZlbClient().RegisterBackend(request)
	if err != nil {
		return diag.FromErr(err)
	}

	// Use format "zlb_id:listener_id" as ID
	d.SetId(*request.LoadBalancerId + ":" + *request.ListenerId)

	return resourceZenlayerCloudZlbBackendRead(ctx, d, meta)
}

func resourceZenlayerCloudZlbBackendRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb_backend.read")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	// Parse zlb_id and listener_id from ID
	items, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	zlbId := items[0]
	listenerId := items[1]

	request := zlb.NewDescribeBackendsRequest()
	request.LoadBalancerId = &zlbId
	request.ListenerId = &listenerId

	response, err := zlbService.client.WithZlbClient().DescribeBackends(request)

	if err != nil {
		ee, ok := err.(*common.ZenlayerCloudSdkError)
		if !ok {
			return diag.FromErr(err)
		}
		if ee.Code == common2.ResourceNotFound || ee.Code == "INVALID_LB_LISTENER_NOT_FOUND" {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Set backends data
	backends := make([]map[string]interface{}, 0)
	for _, backend := range response.Response.Backends {
		backendMap := map[string]interface{}{
			"instance_id":        backend.InstanceId,
			"private_ip_address": backend.PrivateIpAddress,
			"port":               backend.BackendPort,
			"weight":             backend.Weight,
		}
		backends = append(backends, backendMap)
	}

	_ = d.Set("zlb_id", zlbId)
	_ = d.Set("listener_id", listenerId)
	_ = d.Set("backends", backends)

	return nil
}

func resourceZenlayerCloudZlbBackendUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb_backend.update")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	items, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	zlbId := items[0]
	listenerId := items[1]

	if d.HasChange("backends") {
		// Get old and new backends
		oldBackends, newBackends := d.GetChange("backends")

		// Unregister old backends
		oldBackendSet := oldBackends.(*schema.Set)
		newBackendSet := newBackends.(*schema.Set)

		// Find backends to remove
		removeBackends := oldBackendSet.Difference(newBackendSet).List()
		addBackends := newBackendSet.Difference(oldBackendSet).List()
		oldIds := getIdSetFromServers(removeBackends)
		newIds := getIdSetFromServers(addBackends)
		updateSet := oldIds.Intersection(newIds)
		addSet := newIds.Difference(oldIds)
		removeSet := oldIds.Difference(newIds)

		if removeSet.Len() > 0 {

			unregisterRequest := zlb.NewDeregisterBackendRequest()
			unregisterRequest.LoadBalancerId = &zlbId
			unregisterRequest.ListenerId = &listenerId

			backendList := make([]*zlb.BackendServer, 0, len(removeBackends))
			for _, backend := range removeBackends {
				item := backend.(map[string]interface{})
				if !removeSet.Contains(item["instance_id"]) {
					continue
				}
				backendItem := &zlb.BackendServer{
					InstanceId:       common.String(item["instance_id"].(string)),
					PrivateIpAddress: common.String(item["private_ip_address"].(string)),
				}
				backendList = append(backendList, backendItem)
			}
			unregisterRequest.BackendServers = backendList

			_, err := zlbService.client.WithZlbClient().DeregisterBackend(unregisterRequest)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Register new backends
		if addSet.Len()> 0 {
			registerRequest := zlb.NewRegisterBackendRequest()
			registerRequest.LoadBalancerId = &zlbId
			registerRequest.ListenerId = &listenerId

			backendList := make([]*zlb.BackendServer, 0, len(addBackends))
			for _, backend := range addBackends {
				item := backend.(map[string]interface{})
				if !addSet.Contains(item["instance_id"]) {
					continue
				}
				backendItem := &zlb.BackendServer{
					InstanceId:       common.String(item["instance_id"].(string)),
					PrivateIpAddress: common.String(item["private_ip_address"].(string)),
				}

				if weight, exists := item["weight"]; exists {
					v := weight.(int)
					backendItem.Weight = common.Integer(v)
				}

				if port, exists := item["port"]; exists {
					v := port.(int)
					if v != 0 {
						backendItem.Port = common.Integer(v)
					}
				}

				backendList = append(backendList, backendItem)
			}
			registerRequest.BackendServers = backendList

			_, err := zlbService.client.WithZlbClient().RegisterBackend(registerRequest)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// update backend
		if updateSet.Len() > 0 {
			backendServers := make([]*zlb.BackendServer, 0)

			for _, server := range d.Get("backends").(*schema.Set).List() {
				s := server.(map[string]interface{})
				if updateSet.Contains(s["instance_id"]) {
					backendServer := &zlb.BackendServer{
						InstanceId: common.String(s["instance_id"].(string)),
						PrivateIpAddress: common.String(s["private_ip_address"].(string)),
					}

					if weight, exists := s["weight"]; exists {
						v := weight.(int)
						backendServer.Weight = common.Integer(v)
					}

					if port, exists := s["port"]; exists {
						v := port.(int)
						if v != 0 {
							backendServer.Port = common.Integer(v)
						}
					}

					backendServers = append(backendServers, backendServer)
				}
			}
			modifyRequest := zlb.NewModifyBackendRequest()
			modifyRequest.LoadBalancerId = &zlbId
			modifyRequest.ListenerId = &listenerId
			modifyRequest.BackendServers = backendServers

			_, err := zlbService.client.WithZlbClient().ModifyBackend(modifyRequest)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceZenlayerCloudZlbBackendRead(ctx, d, meta)
}

func getIdSetFromServers(items []interface{}) *schema.Set {
	rmId := make([]interface{}, 0)
	for _, item := range items {
		server := item.(map[string]interface{})
		rmId = append(rmId, fmt.Sprintf("%s", server["instance_id"]))
	}
	return schema.NewSet(schema.HashString, rmId)
}

func resourceZenlayerCloudZlbBackendDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb_backend.delete")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	items, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	zlbId := items[0]
	listenerId := items[1]

	// Get all backends to unregister
	backends := d.Get("backends").(*schema.Set).List()
	if len(backends) == 0 {
		return nil
	}

	request := zlb.NewDeregisterBackendRequest()
	request.LoadBalancerId = &zlbId
	request.ListenerId = &listenerId

	backendList := make([]*zlb.BackendServer, 0, len(backends))
	for _, backend := range backends {
		item := backend.(map[string]interface{})
		backendItem := &zlb.BackendServer{
			InstanceId:       common.String(item["instance_id"].(string)),
			PrivateIpAddress: common.String(item["private_ip_address"].(string)),
		}
		backendList = append(backendList, backendItem)
	}

	request.BackendServers = backendList

	_, err = zlbService.client.WithZlbClient().DeregisterBackend(request)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
