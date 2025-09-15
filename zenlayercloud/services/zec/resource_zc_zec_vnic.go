package zec

import (
	"context"
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
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"time"
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
		//stackTypeForceNewFunc(),
		),
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-vNIC",
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description:  "The name of the vNIC. maximum length is 63.",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of a VPC subnet.",
			},
			"stack_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The stack type of the subnet. Valid values: `IPv4`, `IPv6`, `IPv4_IPv6`.",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The ID of a security group. If absent, the security group under VPC will be used.",
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
			// The IPv6 network billing
			"ipv6_internet_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ByBandwidth", "ByTrafficPackage", "BandwidthCluster"}, false),
				ForceNew:     true,
				Description:  "Network billing methods for public IPv6. Valid values: `ByBandwidth`, `ByTrafficPackage`, `BandwidthCluster`.",
			},
			"ipv6_bandwidth": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Bandwidth of public IPv6. Measured in Mbps.",
			},
			"ipv6_bandwidth_cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Bandwidth cluster ID for public IPv6. Required when `ipv6_internet_charge_type` is `BandwidthCluster`.",
			},
			"ipv6_traffic_package_size": {
				Type:         schema.TypeFloat,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.FloatAtLeast(0.0),
				Description:  "Traffic Package size for public IPv6. Measured in TB. Only valid when `ipv6_internet_charge_type` is `ByTrafficPackage`.",
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
			request.Resources = []*string{common.String(nicId)}

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

	return resourceZenlayerCloudZecVNicRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVNicCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vnic.create")()

	request := zec2.NewCreateNetworkInterfaceRequest()
	request.SubnetId = common.String(d.Get("subnet_id").(string))
	request.Name = common.String(d.Get("name").(string))

	// v6
	if v, ok := d.GetOk("ipv6_internet_charge_type"); ok {
		request.InternetChargeType = common.String(v.(string))
	}
	if v, ok := d.GetOk("ipv6_bandwidth"); ok {
		request.Bandwidth = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("ipv6_bandwidth_cluster_id"); ok {
		request.ClusterId = common.String(v.(string))
	}
	if v, ok := d.GetOk("ipv6_traffic_package_size"); ok {
		request.PackageSize = common.Float64(v.(float64))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common.String(v.(string))
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
			return common2.RetryError(ctx, err)
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

	return resourceZenlayerCloudZecVNicRead(ctx, d, meta)
}

func resourceZenlayerCloudZecVNicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	vnicId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var nic *zec2.NicInfo
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
		return nil
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
	_ = d.Set("primary_ipv4", nic.PrimaryIpv4)
	_ = d.Set("primary_ipv6", nic.PrimaryIpv6)
	// TODO
	//_ = d.Set("ipv6_internet_charge_type", nic.InternetChargeType)
	//_ = d.Set("ipv6_bandwidth", nic.Bandwidth)
	//_ = d.Set("ipv6_bandwidth_cluster_id", nic.ClusterId)
	//_ = d.Set("ipv6_traffic_package_size", nic.PackageSize)
	_ = d.Set("resource_group_id", nic.ResourceGroup.ResourceGroupId)
	_ = d.Set("resource_group_name", nic.ResourceGroup.ResourceGroupName)
	_ = d.Set("create_time", nic.CreateTime)
	_ = d.Set("security_group_id", nic.SecurityGroupId)
	return diags

}

func nicIsOperating(status string) bool {
	return common2.IsContains([]string{
		ZecVnicStatusCreating,
		ZecVnicStatusDeleting,
		ZecVnicStatusAttaching,
		ZecVnicStatusDetaching}, status)
}
