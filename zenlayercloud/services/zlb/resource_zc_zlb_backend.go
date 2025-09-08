package zlb

import (
	"context"
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
							Computed:     true,
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
		if v, ok := d.GetOk("port"); ok {
			backendItem.Port = common.Integer(v.(int))
		}
		if v, ok := d.GetOk("weight"); ok {
			backendItem.Weight = common.Integer(v.(int))
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
		if len(removeBackends) > 0 {
			unregisterRequest := zlb.NewDeregisterBackendRequest()
			unregisterRequest.LoadBalancerId = &zlbId
			unregisterRequest.ListenerId = &listenerId

			backendList := make([]*zlb.BackendServer, 0, len(removeBackends))
			for _, backend := range removeBackends {
				item := backend.(map[string]interface{})
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
		addBackends := newBackendSet.Difference(oldBackendSet).List()
		if len(addBackends) > 0 {
			registerRequest := zlb.NewRegisterBackendRequest()
			registerRequest.LoadBalancerId = &zlbId
			registerRequest.ListenerId = &listenerId

			backendList := make([]*zlb.BackendServer, 0, len(addBackends))
			for _, backend := range addBackends {
				item := backend.(map[string]interface{})
				backendItem := &zlb.BackendServer{
					InstanceId:       common.String(item["instance_id"].(string)),
					PrivateIpAddress: common.String(item["private_ip_address"].(string)),
					Port:             common.Integer(item["port"].(int)),
					Weight:           common.Integer(item["weight"].(int)),
				}
				backendList = append(backendList, backendItem)
			}
			registerRequest.BackendServers = backendList

			_, err := zlbService.client.WithZlbClient().RegisterBackend(registerRequest)
			if err != nil {
				return diag.FromErr(err)
			}
		}

		// Handle modified backends (same instance_id and port, but different weight)
		commonBackends := oldBackendSet.Intersection(newBackendSet).List()
		// 根据接口文档，使用ModifyBackend接口直接修改后端服务器权重
		// 接口参数包括loadBalancerId, listenerId, backendServers
		modifyRequest := zlb.NewModifyBackendRequest()
		modifyRequest.LoadBalancerId = &zlbId
		modifyRequest.ListenerId = &listenerId

		// 构造需要修改的后端服务器列表
		backendServers := make([]*zlb.BackendServer, 0)
		for _, backend := range commonBackends {
			item := backend.(map[string]interface{})

			// 检查权重是否发生变化
			oldWeight, newWeight := d.GetChange("backends." + item["instance_id"].(string) + ".weight")
			if oldWeight != newWeight {
				backendServer := &zlb.BackendServer{
					InstanceId: common.String(item["instance_id"].(string)),
					Port:       common.Integer(item["port"].(int)),
					Weight:     common.Integer(newWeight.(int)),
				}
				backendServers = append(backendServers, backendServer)
			}
		}

		// 只有当有后端服务器需要修改时才调用接口
		if len(backendServers) > 0 {
			modifyRequest.BackendServers = backendServers

			_, err := zlbService.client.WithZlbClient().ModifyBackend(modifyRequest)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceZenlayerCloudZlbBackendRead(ctx, d, meta)
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
			InstanceId: common.String(item["instance_id"].(string)),
			Port:       common.Integer(item["port"].(int)),
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
