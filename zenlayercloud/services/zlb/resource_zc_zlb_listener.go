package zlb

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zlb "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zlb20250401"
	"regexp"
	"time"
)

func ResourceZenlayerCloudZlbListener() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZlbListenerCreate,
		ReadContext:   resourceZenlayerCloudZlbListenerRead,
		UpdateContext: resourceZenlayerCloudZlbListenerUpdate,
		DeleteContext: resourceZenlayerCloudZlbListenerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			healthCheckValidFunc(),
			healthCheckHTTPValidFunc(),
			healthCheckDisableValidFunc(),
		),

		Schema: map[string]*schema.Schema{
			"zlb_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of load balancer that the listener belongs to.",
			},
			"listener_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the load balancer listener.",
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The protocol of listener. Valid values: `TCP`, `UDP`.",
				ValidateFunc: validation.StringInSlice([]string{"TCP", "UDP"}, false),
			},
			"port": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The port of listener. Multiple ports are separated by commas. When the port is a range, connect with -, for example: 10000-10005.The value range of the port is 1 to 65535. Please note that the port cannot overlap with other ports of the listener.",
			},
			"scheduler": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "mh",
				Description:  "Scheduling algorithm of the listener. Valid values: `mh`, `rr`, `wrr`, `lc`, `wlc`, `sh`, `dh`. Default value: `mh`.",
				ValidateFunc: validation.StringInSlice([]string{"mh", "rr", "wrr", "lc", "wlc", "sh", "dh"}, false),
			},
			"kind": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "FNAT",
				Description:  "Forwarding mode of the listener. Valid values: `DR`(stands for Direct Routing), `FNAT`(stands for Full NAT), `DNAT`(stands for Destination NAT). Default is `FNAT`.",
				ValidateFunc: validation.StringInSlice([]string{"DR", "FNAT", "DNAT"}, false),
			},
			"health_check_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates whether health check is enabled. Default is `true`. When health check is disabled, other health check parameter can't be set.",
			},
			"health_check_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Health check protocols. Valid values: `PING_CHECK`, `TCP`, `HTTP_GET`.",
				ValidateFunc: validation.StringInSlice([]string{"PING_CHECK", "TCP", "HTTP_GET"}, false),
			},
			"health_check_http_get_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/.*`), "The value should start with '/'"),
				Description:  "HTTP request URL for health check. The value should start with '/', Default is `/`.",
			},
			"health_check_delay_try": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 15),
				Description:  "Health check delay try time.Valid values: `1` to `15`. `health_check_delay_try` takes effect only if `health_check_enabled` is set to true. Default is `2`.",
			},
			"health_check_conn_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 15),
				Description:  "Connection timeout for health check. Valid values: `1` to `15`. `health_check_conn_timeout` takes effect only if `health_check_enabled` is set to true. Default is `2`.",
			},
			"health_check_http_status_code": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(100, 599),
				Description:  "HTTP status code for health check. Required when `check_type` is `HTTP_GET`. Valid values: `100` to `599`.",
			},
			"health_check_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
				Description:  "Health check port. Defaults to the backend server port. Valid values: `1` to `65535`. `health_check_port` takes effect only if `health_check_enabled` is set to true.",
			},
			"health_check_retry": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 5),
				Description:  "Number of retry attempts for health check. Valid values: `1` to `5`. `health_check_retry` takes effect only if `health_check_enabled` is set to true. Default is `2`.",
			},
			"health_check_delay_loop": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(3, 30),
				Description:  "Interval between health checks. Measured in second. Valid values: `3` to `30`. `health_check_delay_loop` takes effect only if `health_check_enabled` is set to true. Default is `3`.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the listener.",
			},
		},
	}
}

func healthCheckHTTPValidFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("health_check_type", func(ctx context.Context, value, meta interface{}) bool {
		return value == "HTTP_GET"
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {

		if _, ok := diff.GetOk("health_check_http_status_code"); !ok {
			return errors.New("`health_check_http_status_code` is missing when `health_check_type` is set to `HTTP_GET`")
		}

		return nil
	})
}

func healthCheckDisableValidFunc() schema.CustomizeDiffFunc {

	return customdiff.IfValue("health_check_enabled", func(ctx context.Context, value, meta interface{}) bool {
		return value == false
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {

		if !diff.GetRawConfig().GetAttr("health_check_http_get_url").IsNull() {
			return errors.New("`health_check_http_get_url` can't be set when `health_check_enabled` is set to `false`")
		}

		if !diff.GetRawConfig().GetAttr("health_check_conn_timeout").IsNull() {
			return errors.New("`health_check_conn_timeout` can't be set when `health_check_enabled` is set to `false`")
		}


		if !diff.GetRawConfig().GetAttr("health_check_http_status_code").IsNull() {
			return errors.New("`health_check_http_status_code` can't be set when `health_check_enabled` is set to `false`")
		}

		if !diff.GetRawConfig().GetAttr("health_check_port").IsNull() {
			return errors.New("`health_check_port` can't be set when `health_check_enabled` is set to `false`")
		}

		if !diff.GetRawConfig().GetAttr("health_check_retry").IsNull() {
			return errors.New("`health_check_retry` can't be set when `health_check_enabled` is set to `false`")
		}

		if !diff.GetRawConfig().GetAttr("health_check_delay_loop").IsNull() {
			return errors.New("`health_check_delay_loop` can't be set when `health_check_enabled` is set to `false`")
		}

		if !diff.GetRawConfig().GetAttr("health_check_type").IsNull() {
			return errors.New("`health_check_type` can't be set when `health_check_enabled` is set to `false`")
		}
		return nil
	})

}

func healthCheckValidFunc() schema.CustomizeDiffFunc {

	return customdiff.IfValue("health_check_enabled", func(ctx context.Context, value, meta interface{}) bool {
		return value == true
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {

		if _, ok := diff.GetOk("health_check_type"); !ok {
			return errors.New("`health_check_type` is missing when `health_check_enabled` is set to `true`")
		}

		return nil
	})

}

func resourceZenlayerCloudZlbListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb_listener.create")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zlb.NewCreateListenerRequest()

	zlbId := d.Get("zlb_id").(string)
	request.LoadBalancerId = &zlbId
	request.ListenerName = common.String(d.Get("listener_name").(string))
	request.Protocol = common.String(d.Get("protocol").(string))
	request.Port = common.String(d.Get("port").(string))
	request.Scheduler = common.String(d.Get("scheduler").(string))
	request.Kind = common.String(d.Get("kind").(string))

	healthCheck := &zlb.HealthCheck{
		Enabled: common.Bool(d.Get("health_check_enabled").(bool)),
	}

	if v, ok := d.GetOk("health_check_type"); ok {
		healthCheck.CheckType = common.String(v.(string))
	}
	if v, ok := d.GetOk("health_check_http_get_url"); ok {
		healthCheck.CheckHttpGetUrl = common.String(v.(string))
	}
	if v, ok := d.GetOk("health_check_delay_try"); ok {
		healthCheck.CheckDelayTry = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("health_check_conn_timeout"); ok {
		healthCheck.CheckConnTimeout = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("health_check_http_status_code"); ok {
		healthCheck.CheckHttpStatusCode = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("health_check_port"); ok {
		healthCheck.CheckPort = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("health_check_retry"); ok {
		healthCheck.CheckRetry = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("health_check_delay_loop"); ok {
		healthCheck.CheckDelayLoop = common.Integer(v.(int))
	}

	request.HealthCheck = healthCheck

	response, err := zlbService.client.WithZlbClient().CreateListener(request)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", zlbId, *response.Response.ListenerId))

	return resourceZenlayerCloudZlbListenerRead(ctx, d, meta)
}

func resourceZenlayerCloudZlbListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb_listener.read")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	// lbId:listenerId
	items, err := common2.ParseResourceId(d.Id(), 2)
	lbId := items[0]
	listenerId := items[1]

	request := zlb.NewDescribeListenersRequest()
	request.LoadBalancerId = &lbId
	request.ListenerIds = []string{listenerId}

	response, err := zlbService.client.WithZlbClient().DescribeListeners(request)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(response.Response.Listeners) < 1 {
		d.SetId("")
		return nil
	}

	listener := response.Response.Listeners[0]

	_ = d.Set("zlb_id", lbId)
	_ = d.Set("listener_name", listener.ListenerName)
	_ = d.Set("protocol", listener.Protocol)
	_ = d.Set("port", listener.Port)
	_ = d.Set("scheduler", listener.Scheduler)
	_ = d.Set("kind", listener.Kind)
	_ = d.Set("create_time", listener.CreateTime)
	_ = d.Set("health_check_enabled", listener.HealthCheck.Enabled)
	_ = d.Set("health_check_type", listener.HealthCheck.CheckType)
	_ = d.Set("health_check_http_get_url", listener.HealthCheck.CheckHttpGetUrl)
	_ = d.Set("health_check_delay_try", listener.HealthCheck.CheckDelayTry)
	_ = d.Set("health_check_conn_timeout", listener.HealthCheck.CheckConnTimeout)
	_ = d.Set("health_check_http_status_code", listener.HealthCheck.CheckHttpStatusCode)
	_ = d.Set("health_check_port", listener.HealthCheck.CheckPort)
	_ = d.Set("health_check_retry", listener.HealthCheck.CheckRetry)
	_ = d.Set("health_check_delay_loop", listener.HealthCheck.CheckDelayLoop)

	return nil
}

func resourceZenlayerCloudZlbListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb_listener.update")()

	items, err := common2.ParseResourceId(d.Id(), 2)
	lbId := items[0]
	listenerId := items[1]

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zlb.NewModifyListenerRequest()
	request.ListenerId = &listenerId
	request.LoadBalancerId = &lbId

	if d.HasChange("listener_name") {
		request.ListenerName = common.String(d.Get("listener_name").(string))
	}

	if d.HasChange("scheduler") {
		request.Scheduler = common.String(d.Get("scheduler").(string))
	}

	if d.HasChange("kind") {
		request.Kind = common.String(d.Get("kind").(string))
	}

	if d.HasChange("port") {
		request.Port = common.String(d.Get("port").(string))
	}

	if d.HasChanges("health_check_enabled", "health_check_type", "health_check_http_get_url",
		"health_check_delay_try", "health_check_conn_timeout", "health_check_http_status_code",
		"health_check_port", "health_check_retry", "health_check_delay_loop") {

		healthCheck := &zlb.HealthCheck{}

		if v, ok := d.GetOk("health_check_enabled"); ok {
			healthCheck.Enabled = common.Bool(v.(bool))
		} else {
			healthCheck.Enabled = common.Bool(false)
		}

		if v, ok := d.GetOk("health_check_type"); ok {
			healthCheck.CheckType = common.String(v.(string))
		}
		if v, ok := d.GetOk("health_check_http_get_url"); ok {
			healthCheck.CheckHttpGetUrl = common.String(v.(string))
		}
		if v, ok := d.GetOk("health_check_delay_try"); ok {
			healthCheck.CheckDelayTry = common.Integer(v.(int))
		}
		if v, ok := d.GetOk("health_check_conn_timeout"); ok {
			healthCheck.CheckConnTimeout = common.Integer(v.(int))
		}
		if v, ok := d.GetOk("health_check_http_status_code"); ok {
			healthCheck.CheckHttpStatusCode = common.Integer(v.(int))
		}
		if v, ok := d.GetOk("health_check_port"); ok {
			healthCheck.CheckPort = common.Integer(v.(int))
		}

		if v, ok := d.GetOk("health_check_retry"); ok {
			healthCheck.CheckRetry = common.Integer(v.(int))
		}
		if v, ok := d.GetOk("health_check_delay_loop"); ok {
			healthCheck.CheckDelayLoop = common.Integer(v.(int))
		}

		request.HealthCheck = healthCheck
	}

	_, err = zlbService.client.WithZlbClient().ModifyListener(request)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceZenlayerCloudZlbListenerRead(ctx, d, meta)
}

func resourceZenlayerCloudZlbListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb_listener.delete")()

	items, err := common2.ParseResourceId(d.Id(), 2)
	lbId := items[0]
	listenerId := items[1]

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zlb.NewDeleteListenerRequest()
	request.ListenerId = &listenerId
	request.LoadBalancerId = &lbId

	_, err = zlbService.client.WithZlbClient().DeleteListener(request)
	if err != nil {
		ee, ok := err.(*common.ZenlayerCloudSdkError)
		if !ok {
			return diag.FromErr(err)
		}
		if ee.Code == "INVALID_LB_LISTENER_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
			// listener doesn't exist
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
