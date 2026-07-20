package zec

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/zrm"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	traffic "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/traffic20240326"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecVNic() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecVNicCreate,
		ReadContext:   resourceZenlayerCloudZecVNicRead,
		UpdateContext: resourceZenlayerCloudZecVNicUpdate,
		DeleteContext: resourceZenlayerCloudZecVNicDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			// ForceNew when ipv6_internet_charge_type changes on an existing IPv6 vNIC.
			// Only triggers when old value is non-empty (IPv6 already assigned), not on first assignment.
			customdiff.ForceNewIfChange("ipv6_internet_charge_type", func(_ context.Context, old, new, _ interface{}) bool {
				return old.(string) != "" && new.(string) != "" && old.(string) != new.(string)
			}),
			// Mark computed fields as unknown when stack_type changes.
			// IPv4_IPv6 -> IPv4: primary_ipv6 and IPv6 billing fields (Optional+Computed) are
			// cleared by Read after apply, but plan won't show them changing without SetNewComputed.
			// IPv4 -> IPv4_IPv6: only primary_ipv6 needs marking; billing fields are set explicitly by user.
			func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
				if d.HasChange("stack_type") {
					_ = d.SetNewComputed("primary_ipv6")
					oldVal, _ := d.GetChange("stack_type")
					if oldVal.(string) == "IPv4_IPv6" {
						_ = d.SetNewComputed("ipv6_internet_charge_type")
						_ = d.SetNewComputed("ipv6_bandwidth")
						_ = d.SetNewComputed("rate_limit_mode")
					}
				}
				return nil
			},
			// ipv6_bandwidth_cluster_id is only valid when ipv6_internet_charge_type = BandwidthCluster.
			func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
				clusterId := d.Get("ipv6_bandwidth_cluster_id").(string)
				chargeType := d.Get("ipv6_internet_charge_type").(string)
				if clusterId != "" && chargeType != "BandwidthCluster" {
					return fmt.Errorf("ipv6_bandwidth_cluster_id can only be set when ipv6_internet_charge_type is BandwidthCluster")
				}
				return nil
			},
			// ipv6_traffic_package_size is only valid when ipv6_internet_charge_type = ByTrafficPackage.
			func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
				if v, ok := d.GetOk("ipv6_traffic_package_size"); ok && v.(float64) > 0 {
					if d.Get("ipv6_internet_charge_type").(string) != "ByTrafficPackage" {
						return fmt.Errorf("ipv6_traffic_package_size can only be set when ipv6_internet_charge_type is `ByTrafficPackage`")
					}
				}
				return nil
			},
		),
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-vNIC",
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description:  "The name of the vNIC. Maximum length is 63. Default is `Terraform-vNIC`.",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of a VPC subnet.",
			},
			"stack_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"IPv4", "IPv6", "IPv4_IPv6"}, false),
				Description:  "The stack type of the vNIC. Valid values: `IPv4`, `IPv6`, `IPv4_IPv6`. Supports in-place update between `IPv4` and `IPv4_IPv6`.",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The ID of a security group. If not specified, the security group attached to the VPC will be used. Ensure the VPC has a security group attached via `zenlayercloud_zec_vpc_security_group_attachment` before applying.",
			},
			"primary_ipv4": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The primary IPv4 address of the vNIC.",
			},
			"primary_ipv6": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The primary IPv6 address of the vNIC.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the vNIC belongs to, default to ID of Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the vNIC belongs to, default to Default Resource Group.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The available tags within this vNIC.",
			},
			// The IPv6 network billing
			"ipv6_internet_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"ByBandwidth", "ByTrafficPackage", "BandwidthCluster"}, false),
				Description:  "Network billing methods for public IPv6. Valid values: `ByBandwidth`, `ByTrafficPackage`, `BandwidthCluster`. Changing this field when a public IPv6 is already assigned forces resource recreation.",
			},
			"ipv6_bandwidth": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Bandwidth of public IPv6. Measured in Mbps.",
			},
			"ipv6_bandwidth_cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Bandwidth cluster ID for public IPv6. Required when `ipv6_internet_charge_type` is `BandwidthCluster`. Must not be set for other charge types.",
			},
			"ipv6_traffic_package_size": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ValidateFunc: validation.FloatAtLeast(0.0),
				Description:  "Traffic Package size for public IPv6. Measured in TB. Only valid when `ipv6_internet_charge_type` is `ByTrafficPackage`. Range: >= 0.",
			},
			"rate_limit_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{ZecEipRateLimitModeLoose, ZecEipRateLimitModeStrict}, false),
				Description:  "Bandwidth rate limit mode for public IPv6. Valid values: `LOOSE`, `STRICT`. Only takes effect on public IPv6 with fixed bandwidth (`ipv6_internet_charge_type` = `ByBandwidth`).",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the vNIC.",
			},
		},
	}
}

func stackTypeForceNewFunc() schema.CustomizeDiffFunc {
	// Valid only from IPv4 -> IPv4_IPv6
	return customdiff.ForceNewIf("stack_type", func(_ context.Context, d *schema.ResourceDiff, _ interface{}) bool {
		oldValue, newValue := d.GetChange("stack_type")
		return oldValue.(string) == "IPv4_IPv6" || (oldValue.(string) == "IPv6" && newValue.(string) == "IPv4_IPv6")
	})
}

func resourceZenlayerCloudZecVNicDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vnicId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteVnicById(ctx, vnicId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound || ee.Code == INVALID_NIC_NOT_FOUND {
				// vpc doesn't exist
				return nil
			}

			return resource.NonRetryableError(errRet)
		}
		return nil
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		vnic, errRet := zecService.DescribeNicById(ctx, vnicId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if vnic == nil {
			return nil
		}

		if *vnic.Status == ZecVnicStatusDeleting {
			//in recycling
			return resource.RetryableError(fmt.Errorf("vnic (%s) is recycling", vnicId))
		}
		return resource.NonRetryableError(fmt.Errorf("vnic status invalid, current status %s", vnic.Status))
	})

	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceZenlayerCloudZecVNicUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	nicId := d.Id()
	if d.HasChanges("name", "security_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := zecService.ModifyVNicAttribute(ctx, nicId, d.Get("name").(string), d.Get("security_group_id").(string))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := user.NewAddResourceResourceGroupRequest()
			request.ResourceGroupId = common.String(d.Get("resource_group_id").(string))
			request.Resources = []string{nicId}

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

	if d.HasChange("tags") {
		zrmService := zrm.NewZrmService(meta.(*connectivity.ZenlayerCloudClient))
		err := zrmService.ModifyResourceTags(ctx, d, nicId)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("stack_type") {
		oldVal, newVal := d.GetChange("stack_type")
		oldType, newType := oldVal.(string), newVal.(string)

		if oldType == "IPv4" && newType == "IPv4_IPv6" {
			var internetChargeType, clusterId, rateLimitMode *string
			var bandwidth *int
			var packageSize *float64
			if v, ok := d.GetOk("ipv6_internet_charge_type"); ok {
				chargeType := v.(string)
				internetChargeType = common.String(chargeType)
				if chargeType == "ByTrafficPackage" {
					// Always send PackageSize for ByTrafficPackage; d.GetOk returns false for 0.0
					packageSize = common.Float64(d.Get("ipv6_traffic_package_size").(float64))
				}
			}
			if v, ok := d.GetOk("ipv6_bandwidth"); ok {
				bandwidth = common.Integer(v.(int))
			}
			if v, ok := d.GetOk("ipv6_bandwidth_cluster_id"); ok {
				clusterId = common.String(v.(string))
			}
			if v, ok := d.GetOk("rate_limit_mode"); ok {
				rateLimitMode = common.String(v.(string))
			}
			err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
				if errRet := zecService.AssignNetworkInterfaceIpv6(ctx, nicId, internetChargeType, bandwidth, packageSize, clusterId, rateLimitMode); errRet != nil {
					return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
				}
				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}
		} else if oldType == "IPv4_IPv6" && newType == "IPv4" {
			err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
				if errRet := zecService.UnassignNetworkInterfaceIpv6(ctx, nicId); errRet != nil {
					return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
				}
				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}
		} else {
			return diag.FromErr(fmt.Errorf("unsupported stack_type transition: %s -> %s", oldType, newType))
		}
	} else if d.HasChanges("rate_limit_mode", "ipv6_bandwidth", "ipv6_traffic_package_size", "ipv6_bandwidth_cluster_id") {
		address, err := zecService.DescribeNetworkInterfacePublicIPv6ByNicId(ctx, nicId)
		if err != nil {
			return diag.FromErr(err)
		}
		if address == nil || address.Ipv6CidrId == nil {
			return diag.FromErr(fmt.Errorf("public IPv6 not found on vNIC %s", nicId))
		}
		if d.HasChange("ipv6_bandwidth_cluster_id") {
			req := traffic.NewMigrateBandwidthClusterResourcesRequest()
			req.ResourceIdList = []string{*address.Ipv6CidrId}
			req.TargetBandwidthClusterId = common.String(d.Get("ipv6_bandwidth_cluster_id").(string))
			_, err := zecService.client.WithTrafficClient().MigrateBandwidthClusterResources(req)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if d.HasChange("ipv6_bandwidth") {
			if err = zecService.ModifyNetworkInterfaceIpv6Bandwidth(ctx, *address.Ipv6CidrId, d.Get("ipv6_bandwidth").(int)); err != nil {
				return diag.FromErr(err)
			}
			err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
				addr, errRet := zecService.DescribeNetworkInterfacePublicIPv6ByNicId(ctx, nicId)
				if errRet != nil {
					return common2.RetryError(ctx, errRet, common2.InternalServerError)
				}
				if addr == nil {
					return resource.NonRetryableError(fmt.Errorf("public IPv6 not found on vNIC %s while waiting for bandwidth update", nicId))
				}
				if addr.OperationInfo != nil {
					if addr.OperationInfo.Status != nil && *addr.OperationInfo.Status == ZecEipOperationStatusFailed {
						return resource.NonRetryableError(fmt.Errorf("vNIC (%s) ipv6_bandwidth update failed", nicId))
					}
					return resource.RetryableError(fmt.Errorf("vNIC (%s) ipv6_bandwidth is still updating", nicId))
				}
				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if d.HasChange("ipv6_traffic_package_size") && d.Get("ipv6_internet_charge_type").(string) == "ByTrafficPackage" {
			if err = zecService.ModifyNetworkInterfaceIpv6TrafficPackage(ctx, *address.Ipv6CidrId, d.Get("ipv6_traffic_package_size").(float64)); err != nil {
				return diag.FromErr(err)
			}
			err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
				addr, errRet := zecService.DescribeNetworkInterfacePublicIPv6ByNicId(ctx, nicId)
				if errRet != nil {
					return common2.RetryError(ctx, errRet, common2.InternalServerError)
				}
				if addr == nil {
					return resource.NonRetryableError(fmt.Errorf("public IPv6 not found on vNIC %s while waiting for traffic package update", nicId))
				}
				if addr.OperationInfo != nil {
					if addr.OperationInfo.Status != nil && *addr.OperationInfo.Status == ZecEipOperationStatusFailed {
						return resource.NonRetryableError(fmt.Errorf("vNIC (%s) ipv6_traffic_package_size update failed", nicId))
					}
					return resource.RetryableError(fmt.Errorf("vNIC (%s) ipv6_traffic_package_size is still updating", nicId))
				}
				return nil
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if d.HasChange("rate_limit_mode") {
			if err = zecService.ModifyNetworkInterfacePublicIPv6BandwidthLimitMode(ctx, *address.Ipv6CidrId, d.Get("rate_limit_mode").(string)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceZenlayerCloudZecVNicRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVNicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic.create")()

	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}

	request := zec.NewCreateNetworkInterfaceRequest()
	request.SubnetId = common.String(d.Get("subnet_id").(string))
	request.Name = common.String(d.Get("name").(string))

	// v6
	if v, ok := d.GetOk("ipv6_internet_charge_type"); ok {
		chargeType := v.(string)
		request.InternetChargeType = common.String(chargeType)
		if chargeType == "ByTrafficPackage" {
			// Always send PackageSize for ByTrafficPackage; d.GetOk returns false for 0.0
			request.PackageSize = common.Float64(d.Get("ipv6_traffic_package_size").(float64))
		}
	}
	if v, ok := d.GetOk("ipv6_bandwidth"); ok {
		request.Bandwidth = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("ipv6_bandwidth_cluster_id"); ok {
		request.ClusterId = common.String(v.(string))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common.String(v.(string))
	}

	if tags := common2.GetTags(d, "tags"); len(tags) > 0 {
		request.Tags = &zec.TagAssociation{}
		for k, v := range tags {
			tmpKey := k
			tmpValue := v
			request.Tags.Tags = append(request.Tags.Tags, &zec.Tag{
				Key:   &tmpKey,
				Value: &tmpValue,
			})
		}
	}

	if v, ok := d.GetOk("stack_type"); ok {
		request.NicStackType = common.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_id"); ok {
		request.SecurityGroupId = common.String(v.(string))
	}
	vnicId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithZec2Client().CreateNetworkInterface(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create vNIC.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err, common2.OperationTimeout)
		}

		tflog.Info(ctx, "Create vNIC success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.NicId == nil {
			err = fmt.Errorf("vnic id is nil")
			return resource.NonRetryableError(err)
		}
		vnicId = *response.Response.NicId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(vnicId)

	if v, ok := d.GetOk("rate_limit_mode"); ok {
		mode := v.(string)
		err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			address, errRet := zecService.DescribeNetworkInterfacePublicIPv6ByNicId(ctx, vnicId)
			if errRet != nil {
				return common2.RetryError(ctx, errRet)
			}
			if address == nil || address.Ipv6CidrId == nil {
				return resource.RetryableError(fmt.Errorf("waiting for public IPv6 on vNIC %s", vnicId))
			}
			if errRet = zecService.ModifyNetworkInterfacePublicIPv6BandwidthLimitMode(ctx, *address.Ipv6CidrId, mode); errRet != nil {
				return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecVNicRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVNicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	vnicId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var nic *zec.NicInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		nic, errRet = zecService.DescribeNicById(ctx, vnicId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		if nic != nil && nicIsOperating(*nic.Status) {
			return resource.RetryableError(fmt.Errorf("waiting for nic %s operation", nic.NicId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if nic == nil {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The vNIC is not exist",
			Detail:   fmt.Sprintf("The vNIC %s is not exist", vnicId),
		})
		return diags
	}

	if *nic.Status == ZecVnicStatusCreateFailed {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The status of vNIC is `failed`",
			Detail:   fmt.Sprintf("The status of vNIC `%s` is `failed`", vnicId),
		})
		return diags
	}

	// nic info
	_ = d.Set("subnet_id", nic.SubnetId)
	_ = d.Set("name", nic.Name)
	_ = d.Set("stack_type", nic.NicSubnetType)
	_ = d.Set("primary_ipv4", nic.PrimaryIpv4)
	_ = d.Set("primary_ipv6", nic.PrimaryIpv6)
	_ = d.Set("resource_group_id", nic.ResourceGroup.ResourceGroupId)
	_ = d.Set("resource_group_name", nic.ResourceGroup.ResourceGroupName)
	_ = d.Set("create_time", nic.CreateTime)
	_ = d.Set("security_group_id", nic.SecurityGroupId)

	toMap, errRet := common2.TagsToMap(nic.Tags)
	if errRet != nil {
		return diag.FromErr(errRet)
	}
	_ = d.Set("tags", toMap)

	address, errRet := zecService.DescribeNetworkInterfacePublicIPv6ByNicId(ctx, vnicId)
	if errRet != nil {
		return diag.FromErr(errRet)
	}
	if address != nil {
		if address.InternetChargeType != nil {
			_ = d.Set("ipv6_internet_charge_type", address.InternetChargeType)
		}
		if address.Bandwidth != nil {
			_ = d.Set("ipv6_bandwidth", address.Bandwidth)
		}
		if address.BandwidthCluster != nil && address.BandwidthCluster.BandwidthClusterId != nil {
			_ = d.Set("ipv6_bandwidth_cluster_id", address.BandwidthCluster.BandwidthClusterId)
		}
		if address.TrafficPackageSize != nil {
			_ = d.Set("ipv6_traffic_package_size", address.TrafficPackageSize)
		} else {
			_ = d.Set("ipv6_traffic_package_size", 0.0)
		}
		_ = d.Set("rate_limit_mode", address.RateLimitMode)
	} else {
		_ = d.Set("ipv6_internet_charge_type", "")
		_ = d.Set("ipv6_bandwidth", 0)
		_ = d.Set("ipv6_bandwidth_cluster_id", "")
		_ = d.Set("ipv6_traffic_package_size", 0.0)
		_ = d.Set("rate_limit_mode", "")
	}

	return diags

}

func nicIsOperating(status string) bool {
	return common2.IsContains([]string{
		ZecVnicStatusCreating,
		ZecVnicStatusDeleting,
		ZecVnicStatusAttaching,
		ZecVnicStatusDetaching}, status)
}
