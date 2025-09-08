/*
Use this data source to query Load Balancer Backends.

Example Usage
*/
package zlb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zlb "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zlb20250401"
	"time"
)

func DataSourceZenlayerCloudZlbBackends() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZlbBackendsRead,

		Schema: map[string]*schema.Schema{
			"zlb_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of load balancer that the backends belong to.",
			},
			"listener_id": {
				Type:        schema.TypeString,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "The ID of the listener that the backends belong to.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"backends": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of backend servers. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the backend server.",
						},
						"listener_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the listener that the backend server belongs to.",
						},
						"listener_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the listener that the backend server belongs to.",
						},
						"private_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Private IP address of the network interface attached to the instance.",
						},
						"backend_port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Target port for request forwarding and health checks. If left empty, it will follow the listener's port configuration.",
						},
						"listener_port": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Listening port. Use commas (,) to separate multiple ports.Use a hyphen (-) to define a port range, e.g., 10000-10005.",
						},
						"weight": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Forwarding weight of the backend server.",
						},
						"protocol": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Protocol of the backend server.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZlbBackendsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zlb_backends.read")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zlb.NewDescribeBackendsRequest()

	// 设置负载均衡器ID（必填）
	request.LoadBalancerId = common.String(d.Get("zlb_id").(string))
	if v, ok := d.GetOk("listener_id"); ok {
		request.ListenerId = common.String(v.(string))
	}

	var backends []*zlb.ListenerBackend

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		response, err := zlbService.client.WithZlbClient().DescribeBackends(request)
		if err != nil {
			ee, ok := err.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, err)
			} else {
				if ee.Code == common2.ResourceNotFound {
					// backends 空数组
					backends = []*zlb.ListenerBackend{}
					return nil
				}
			}
		} else {
			backends = response.Response.Backends
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	backendList := make([]map[string]interface{}, 0, len(backends))
	ids := make([]string, 0, len(backends))

	for _, backend := range backends {
		// 根据名称正则表达式过滤

		mapping := map[string]interface{}{
			"instance_id":   backend.InstanceId,
			"private_ip":    backend.PrivateIpAddress,
			"protocol":      backend.Protocol,
			"listener_id":   backend.ListenerId,
			"listener_name": backend.ListenerName,
			"listener_port": backend.ListenerPort,
			"backend_port":  backend.BackendPort,
			"weight":        backend.Weight,
		}

		backendList = append(backendList, mapping)

		//instance 如果是nil, 则为-
		var instanceId string

		if backend.InstanceId == nil {
			instanceId = "-"
		} else {
			instanceId = *backend.InstanceId
		}

		// instanceId + "#" + backend.PrivateIpAddress, instance 如果是nil, 则为-
		ids = append(ids, *backend.ListenerId+"#"+instanceId+"#"+*backend.PrivateIpAddress)

	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("backends", backendList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), backendList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
