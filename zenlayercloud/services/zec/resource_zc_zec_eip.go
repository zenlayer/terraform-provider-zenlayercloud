package zec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	traffic "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/traffic20240326"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
)

func ResourceZenlayerCloudZecElasticIP() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecElasticIPCreate,
		ReadContext:   resourceZenlayerCloudZecElasticIPRead,
		UpdateContext: resourceZenlayerCloudZecElasticIPUpdate,
		DeleteContext: resourceZenlayerCloudZecElasticIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			bandwidthClusterIdValidFunc(),
		),
		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The region ID that the elastic IP locates at.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the elastic IP.",
			},
			"ip_network_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"cidr_id"},
				ValidateFunc:  validation.StringInSlice([]string{"BGPLine", "CN2Line", "LocalLine", "ChinaTelecom", "ChinaUnicom", "ChinaMobile", "Cogent"}, false),
				Description:   "Network types of public IPv4. Valid values: `CN2Line`, `LocalLine`, `ChinaTelecom`, `ChinaUnicom`, `ChinaMobile`, `Cogent`.",
			},
			"internet_charge_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ByBandwidth", "ByTrafficPackage", "BandwidthCluster"}, false),
				Description:  "Network billing methods. Valid values: `ByBandwidth`, `ByTrafficPackage`, `BandwidthCluster`.",
			},
			"bandwidth": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Bandwidth. Measured in Mbps.",
			},
			"flow_package_size": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.FloatAtLeast(0),
				Description:  "The Data transfer package. Measured in TB.",
			},
			"cidr_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "CIDR ID, the elastic ip will allocated from given CIDR.",
				ConflictsWith: []string{"ip_network_type"},
			},
			"peer_region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Remote region ID.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Resource group ID.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Name of resource group.",
			},
			"bandwidth_cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Bandwidth cluster ID. Required when `internet_charge_type` is `BandwidthCluster`.",
			},
			"public_ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The elastic ipv4 address.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the elastic IP.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the elastic IP.",
			},
		},
	}
}

func bandwidthClusterIdValidFunc() schema.CustomizeDiffFunc {
	return customdiff.IfValue("internet_charge_type", func(ctx context.Context, value, meta interface{}) bool {
		return value == "BandwidthCluster"
	}, func(ctx context.Context, diff *schema.ResourceDiff, i interface{}) error {
		if _, ok := diff.GetOk("bandwidth_cluster_id"); ok {
			return fmt.Errorf("bandwidth_cluster_id must be set as the internet_charge_type of instance is `BandwidthCluster`")
		}
		return nil
	})

}

func resourceZenlayerCloudZecElasticIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_eip.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec.NewCreateEipsRequest()
	request.RegionId = d.Get("region_id").(string)
	request.Name = d.Get("name").(string)
	request.InternetChargeType = d.Get("internet_charge_type").(string)
	request.EipV4Type = d.Get("ip_network_type").(string)

	if v, ok := d.GetOk("bandwidth"); ok {
		request.Bandwidth = v.(int)
	}

	if v, ok := d.GetOk("flow_package_size"); ok {
		request.FlowPackage = v.(float64)
	}

	if v, ok := d.GetOk("cidr_id"); ok {
		request.CidrId = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("cluster_id"); ok {
		request.ClusterId = v.(string)
	}

	if v, ok := d.GetOk("peer_region_id"); ok {
		request.PeerRegionId = v.(string)
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZecClient().CreateEips(request)
		if err != nil {
			return common.RetryError(ctx, err)
		}

		if len(response.Response.EipIds) == 0 {
			return resource.NonRetryableError(fmt.Errorf("failed to get EIP ID from response"))
		}

		d.SetId(response.Response.EipIds[0])
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceZenlayerCloudZecElasticIPRead(ctx, d, meta)
}

func resourceZenlayerCloudZecElasticIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_eip.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	eipId := d.Id()
	// 等待创建完成

	var eip *zec.EipInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		eipAddress, errRet := zecService.DescribeEipById(ctx, eipId)
		if errRet != nil {
			return common.RetryError(ctx, errRet)
		}

		if eipAddress != nil && ipIsOperating(eipAddress.Status) {
			return resource.RetryableError(fmt.Errorf("waiting for eip %s operation", eipAddress.EipId))
		}
		eip = eipAddress
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if eip == nil || eip.Status == ZecEipStatusCreateFailed ||
		eip.Status == ZecEipStatusRecycle {
		d.SetId("")
		tflog.Info(ctx, "zec eip not exist or created failed or recycled", map[string]interface{}{
			"eipId": eipId,
		})
		return nil
	}

	_ = d.Set("region_id", eip.RegionId)
	_ = d.Set("name", eip.Name)
	_ = d.Set("internet_charge_type", eip.InternetChargeType)
	_ = d.Set("ip_network_type", eip.EipV4Type)
	_ = d.Set("bandwidth", eip.Bandwidth)
	_ = d.Set("flow_package", eip.FlowPackage)
	_ = d.Set("cidr_id", eip.CidrId)
	_ = d.Set("public_ip_address", eip.PublicIpAddresses[0])
	_ = d.Set("resource_group_id", eip.ResourceGroupId)
	_ = d.Set("resource_group_name", eip.ResourceGroupName)
	_ = d.Set("peer_region_id", eip.PeerRegionId)
	_ = d.Set("status", eip.Status)
	if eip.BandwidthCluster != nil {
		_ = d.Set("bandwidth_cluster_id", eip.BandwidthCluster.BandwidthClusterId)
	}
	_ = d.Set("create_time", eip.CreateTime)

	return nil
}

func ipIsOperating(status string) bool {
	return common.IsContains([]string{ZecEipStatusCreating, ZecEipStatusDeleting, ZecEipStatusRecycling}, status)
}

func resourceZenlayerCloudZecElasticIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_eip.update")()
	//
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	d.Partial(true)
	eipId := d.Id()

	if d.HasChange("name") {
		request := zec.NewModifyEipAttributeRequest()
		request.EipId = common2.String(eipId)
		request.Name = common2.String(d.Get("name").(string))
		_, err := zecService.client.WithZecClient().ModifyEipAttribute(request)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	//
	if d.HasChange("bandwidth") {
		request := zec.NewModifyEipBandwidthRequest()
		request.EipId = common2.String(eipId)
		request.Bandwidth = common2.Integer(d.Get("bandwidth").(int))
		_, err := zecService.client.WithZecClient().ModifyEipBandwidth(request)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("bandwidth_cluster_id") {
		request := traffic.NewMigrateBandwidthClusterResourcesRequest()
		request.ResourceIdList = []string{eipId}
		request.TargetBandwidthClusterId = common2.String(d.Get("bandwidth_cluster_id").(string))
		_, err := zecService.client.WithTrafficClient().MigrateBandwidthClusterResources(request)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := user.NewAddResourceResourceGroupRequest()
			request.ResourceGroupId = common2.String(d.Get("resource_group_id").(string))
			request.Resources = []*string{common2.String(eipId)}

			_, err := meta.(*connectivity.ZenlayerCloudClient).WithUsrClient().AddResourceResourceGroup(request)
			if err != nil {
				return common.RetryError(ctx, err, common.InternalServerError, common2.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)

	return resourceZenlayerCloudZecElasticIPRead(ctx, d, meta)
}

func resourceZenlayerCloudZecElasticIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_eip.delete")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec.NewDeleteEipRequest()
	eipId := d.Id()
	request.EipId = eipId

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := zecService.client.WithZecClient().DeleteEip(request)
		if err != nil {
			if sdkError, ok := err.(*common2.ZenlayerCloudSdkError); ok {
				if sdkError.Code == common.ResourceNotFound {
					return nil
				}
			}
			return common.RetryError(ctx, err)
		}
		return nil
	})

	notExist := false

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		eip, errRet := zecService.DescribeEipById(ctx, eipId)
		if errRet != nil {
			return common.RetryError(ctx, errRet, common.InternalServerError)
		}
		if eip == nil {
			notExist = true
			return nil
		}

		if eip.Status == ZecEipStatusRecycle {
			return nil
		}
		if eip.Status == ZecEipStatusRecycling {
			return resource.RetryableError(fmt.Errorf("zec eip (%s) is recycling", eipId))
		}
		if eip.Status == ZecEipStatusDeleting {
			return resource.RetryableError(fmt.Errorf("zec eip (%s) is deleting", eipId))
		}

		return resource.NonRetryableError(fmt.Errorf("zec eip status is not recycle, current status %s", eip.Status))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if notExist {
		return nil
	}

	tflog.Debug(ctx, "Releasing zec EIP ...", map[string]interface{}{
		"eipId": eipId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, err := zecService.client.WithZecClient().DeleteEip(request)
		if err != nil {
			if sdkError, ok := err.(*common2.ZenlayerCloudSdkError); ok {
				if sdkError.Code == common.ResourceNotFound {
					return nil
				}
			}
			return common.RetryError(ctx, err)
		}

		return nil
	})

	return resourceZenlayerCloudZecElasticIPRead(ctx, d, meta)
}
