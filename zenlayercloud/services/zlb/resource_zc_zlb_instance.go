package zlb

import (
	"context"
	"fmt"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	zlb "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zlb20250401"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

func ResourceZenlayerCloudZlbInstance() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZlbInstanceCreate,
		ReadContext:   resourceZenlayerCloudZlbInstanceRead,
		UpdateContext: resourceZenlayerCloudZlbInstanceUpdate,
		DeleteContext: resourceZenlayerCloudZlbInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of region that the load balancer instance locates at.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of VPC that the load balancer instance belongs to.",
			},
			"zlb_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-ZLB",
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The name of the load balancer instance.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the load balancer instance.",
			},
			"zlb_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the load balancer instance.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the load balancer belongs to, default to Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the load balancer belongs to.",
			},
			"private_ip_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Private virtual Ipv4 addresses of the load balancer instance.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"public_ip_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Public IPv4 addresses(EIP) of the load balancer instance.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceZenlayerCloudZlbInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb.create")()

	request := zlb.NewCreateLoadBalancerRequest()
	request.RegionId = common.String(d.Get("region_id").(string))
	request.VpcId = common.String(d.Get("vpc_id").(string))
	request.LoadBalancerName = common.String(d.Get("zlb_name").(string))
	//request.IpNetworkType = common.String(d.Get("ip_network_type").(string))
	//request.BandwidthMbps = common.Integer(d.Get("bandwidth_mbps").(int))
	//
	//if v, ok := d.GetOk("bandwidth_cluster_id"); ok {
	//	request.BandwidthClusterId = common.String(v.(string))
	//}
	//
	//if v, ok := d.GetOk("traffic_package_size"); ok {
	//	request.TrafficPackageSize = common.Float64(v.(float64))
	//}
	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common.String(v.(string))
	}

	var zlbId string

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithZlbClient().CreateLoadBalancer(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create load balancer instance.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create load balancer instance success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if len(response.Response.LoadBalancerIds) < 1 {
			err = fmt.Errorf("load balancer instance id is nil")
			return resource.NonRetryableError(err)
		}
		zlbId = response.Response.LoadBalancerIds[0]

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(zlbId)

	return resourceZenlayerCloudZlbInstanceRead(ctx, d, meta)
}

func resourceZenlayerCloudZlbInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb.read")()

	var diags diag.Diagnostics

	zlbId := d.Id()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var zlb *zlb.LoadBalancer
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		zlb, errRet = zlbService.DescribeZlbInstanceById(ctx, zlbId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		if zlb != nil && zlbInstanceIsOperating(*zlb.Status) {
			return resource.RetryableError(fmt.Errorf("waiting for load balancer instance %s operation", *zlb.LoadBalancerId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	if *zlb.Status == lbInstanceStatusCreateFailed {
		return diag.Errorf("load balancer `%s`created failed", zlbId)
	}
	if zlb == nil {
		d.SetId("")
		tflog.Info(ctx, "load balancer instance not exist", map[string]interface{}{
			"zlbId": zlbId,
		})
		return nil
	}

	// instance info
	_ = d.Set("region_id", zlb.RegionId)
	_ = d.Set("vpc_id", zlb.VpcId)
	_ = d.Set("zlb_name", zlb.LoadBalancerName)
	_ = d.Set("create_time", zlb.CreateTime)
	_ = d.Set("zlb_status", zlb.Status)
	_ = d.Set("public_ip_address", zlb.PublicIpAddress)
	_ = d.Set("private_ip_address", zlb.PrivateIpAddress)
	_ = d.Set("resource_group_id", zlb.ResourceGroup.ResourceGroupId)
	_ = d.Set("resource_group_name", zlb.ResourceGroup.ResourceGroupName)

	return diags
}

func resourceZenlayerCloudZlbInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb.update")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	zlbId := d.Id()

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := user.NewAddResourceResourceGroupRequest()
			request.ResourceGroupId = common.String(d.Get("resource_group_id").(string))
			request.Resources = []*string{common.String(zlbId)}

			_, err := meta.(*connectivity.ZenlayerCloudClient).WithUsrClient().AddResourceResourceGroup(request)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("zlb_name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := zlbService.ModifyLoadBalancerName(ctx, zlbId, d.Get("zlb_name").(string))

			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZlbInstanceRead(ctx, d, meta)
}

func resourceZenlayerCloudZlbInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zlb.delete")()

	zlbId := d.Id()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zlbService.DeleteZlbInstanceById(ctx, zlbId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	notExist := false

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		instance, errRet := zlbService.DescribeZlbInstanceById(ctx, zlbId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if ok {
				if ee.Code == common2.ResourceNotFound {
					return nil
				}
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if instance == nil {
			notExist = true
			return nil
		}

		if zlbInstanceIsOperating(*instance.Status) {
			return resource.RetryableError(fmt.Errorf("waiting for load balancer instance %s deleting, current status: %s", *instance.LoadBalancerId, *instance.Status))
		}

		if *instance.Status == lbInstanceStatusRecycle {
			return nil
		}

		return resource.NonRetryableError(fmt.Errorf("load balancer instance status is not deleted, current status %s", *instance.Status))
	})

	if err != nil {
		return diag.FromErr(err)
	}
	if !notExist {
		return resourceZenlayerCloudZlbInstanceDelete(ctx, d, meta)
	}

	return nil
}

func zlbInstanceIsOperating(status string) bool {
	return common2.IsContains([]string{
		lbInstanceStatusReleasing, lbInstanceStatusCreating}, status)
}
