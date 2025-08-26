package zlb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zlb "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zlb20250401"
	"regexp"
)

func DataSourceZenlayerCloudZlbListeners() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZlbListenersRead,

		Schema: map[string]*schema.Schema{
			"zlb_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of load balancer that the listeners belong to.",
			},
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the load balancer listeners to be queried.",
			},
			"protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The protocol of listeners to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to filter results by listener name.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"listeners": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of listeners. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"listener_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the load balancer listener.",
						},
						"listener_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the load balancer listener.",
						},
						"protocol": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The protocol of listener.",
						},
						"port": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The port of listener. Use commas (,) to separate multiple ports. Use a hyphen (-) to define a port range, e.g., 10000-10005.",
						},
						"scheduler": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Scheduling algorithm of the listener.",
						},
						"kind": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Forwarding mode of the listener. Valid values: `DR`, `FNAT`.",
						},

						"health_check_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether health check is enabled.",
						},
						"health_check_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Health check protocols.",
						},
						"health_check_http_get_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "HTTP request URL for health check.",
						},
						"health_check_delay_try": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Health check delay try time.",
						},
						"health_check_conn_timeout": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Connection timeout for health check.",
						},
						"health_check_http_status_code": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "HTTP status code for health check.",
						},
						"health_check_port": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Health check port. Defaults to the backend server port.",
						},
						"health_check_retry": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of retry attempts for health check.",
						},
						"health_check_delay_loop": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Interval between health checks. Measured in second.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the listener.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZlbListenersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zlb_listeners.read")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zlb.NewDescribeListenersRequest()
	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			request.ListenerIds = common2.ToStringList(ids)
		}
	}
	request.LoadBalancerId = common.String(d.Get("zlb_id").(string))

	if v, ok := d.GetOk("protocol"); ok {
		request.Protocol = common.String(v.(string))
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	response, err := zlbService.client.WithZlbClient().DescribeListeners(request)

	if err != nil {
		return diag.FromErr(err)
	}
	listeners := response.Response.Listeners
	listenerList := make([]map[string]interface{}, 0, len(listeners))
	ids := make([]string, 0, len(listeners))
	for _, listener := range listeners {
		if nameRegex != nil && !nameRegex.MatchString(*listener.ListenerName) {
			continue
		}
		mapping := map[string]interface{}{
			"listener_id":                   listener.ListenerId,
			"listener_name":                 listener.ListenerName,
			"protocol":                      listener.Protocol,
			"port":                          listener.Port,
			"scheduler":                     listener.Scheduler,
			"kind":                          listener.Kind,
			"create_time":                   listener.CreateTime,
			"health_check_enabled":          listener.HealthCheck.Enabled,
			"health_check_type":             listener.HealthCheck.CheckType,
			"health_check_http_get_url":     listener.HealthCheck.CheckHttpGetUrl,
			"health_check_delay_try":        listener.HealthCheck.CheckDelayTry,
			"health_check_conn_timeout":     listener.HealthCheck.CheckConnTimeout,
			"health_check_http_status_code": listener.HealthCheck.CheckHttpStatusCode,
			"health_check_port":             listener.HealthCheck.CheckPort,
			"health_check_retry":            listener.HealthCheck.CheckRetry,
			"health_check_delay_loop":       listener.HealthCheck.CheckDelayLoop,
		}

		listenerList = append(listenerList, mapping)
		ids = append(ids, *listener.ListenerId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("listeners", listenerList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), listenerList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
