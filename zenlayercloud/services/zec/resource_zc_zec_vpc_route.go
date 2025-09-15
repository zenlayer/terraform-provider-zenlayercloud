package zec

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudGlobalVpcRoute() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudGlobalVpcRouteCreate,
		ReadContext:   resourceZenlayerCloudGlobalVpcRouteRead,
		UpdateContext: resourceZenlayerCloudGlobalVpcRouteUpdate,
		DeleteContext: resourceZenlayerCloudGlobalVpcRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			sourceIpValidFunc(),
		),
		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the VPC.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description:  "The name of the VPC route. The name should start and end with a number or a letter, containing 2 to 63 characters. Only letters, numbers, - and periods (.) are supported.",
			},
			"destination_cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				// TODO IPv6
				ValidateFunc: common2.ValidateCIDRNetworkAddress,
				Description:  "Destination address block.",
			},
			"ip_version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"IPv4", "IPv6"}, false),
				ForceNew:     true,
				Description:  "IP stack type. Valid values: `IPv4`, `IPv6`.",
			},
			"route_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"RouteTypeStatic", "RouteTypePolicy"}, false),
				Description:  "Route type. Valid values: `RouteTypeStatic`, `RouteTypePolicy`.",
			},
			"source_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The source IP matched. Required when the `route_type` is `RouteTypePolicy`.",
			},
			"priority": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Priority of the route entry. Valid value: from `0` to `65535`.",
			},
			"next_hop_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of next hop instance. Currently only ID of vNIC is valid.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the VPC route.",
			},
		},
	}
}

func sourceIpValidFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("route_type", func(ctx context.Context, value, meta interface{}) bool {
		return value == "RouteTypePolicy"
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		if _, ok := diff.GetOk("source_ip"); !ok {
			return errors.New("`source_ip` is only required when `route_type` is `RouteTypePolicy`")
		}

		return nil
	})
}

func resourceZenlayerCloudGlobalVpcRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc_route.delete")()

	routeId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteVpcRoute(ctx,  routeId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound || ee.Code == INVALID_VPC_ROUTE_NOT_FOUND {
				// vpc doesn't exist
				return nil
			}

			return resource.NonRetryableError(errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudGlobalVpcRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc_route.update")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	routeId := d.Id()
	if d.HasChanges("name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {

			err := zecService.ModifyRouteAttribute(ctx,routeId, d.Get("name").(string));
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudGlobalVpcRouteRead(ctx, d, meta)
}

func resourceZenlayerCloudGlobalVpcRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc_route.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zec.NewCreateRouteRequest()
	request.VpcId = common.String(d.Get("vpc_id").(string))
	request.Name =  common.String(d.Get("name").(string))
	request.DestinationCidrBlock =  common.String(d.Get("destination_cidr_block").(string))
	request.IpVersion =  common.String(d.Get("ip_version").(string))
	request.RouteType =  common.String(d.Get("route_type").(string))
	request.Priority = common.Integer(d.Get("priority").(int))
	request.NextHopId =  common.String(d.Get("next_hop_id").(string))
	if v, ok := d.GetOk("source_ip"); ok {
		request.SourceIp =  common.String(v.(string))
	}

	routeId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZecClient().CreateRoute(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create vpc route.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create global vpc route success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.RouteId == "" {
			err = fmt.Errorf("routeId id is nil")
			return resource.NonRetryableError(err)
		}
		routeId = response.Response.RouteId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(routeId)
	return resourceZenlayerCloudGlobalVpcRouteRead(ctx, d, meta)
}

func resourceZenlayerCloudGlobalVpcRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc_route.read")()

	var diags diag.Diagnostics

	routeId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var route *zec.RouteInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		route, errRet = zecService.DescribeVpcRouteById(ctx, routeId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if route == nil {
		d.SetId("")
		return nil
	}

	_ = d.Set("vpc_id", route.VpcId)
	_ = d.Set("name", route.Name)
	_ = d.Set("destination_cidr_block", route.DestinationCidrBlock)
	_ = d.Set("ip_version", route.IpVersion)
	_ = d.Set("route_type", route.Type)
	_ = d.Set("source_ip", route.SourceIp)
	_ = d.Set("priority", route.Priority)
	_ = d.Set("next_hop_id", route.NextHopId)
	_ = d.Set("create_time", route.CreateTime)

	return diags

}
