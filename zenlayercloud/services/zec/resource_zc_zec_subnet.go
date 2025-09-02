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
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecSubnetCreate,
		ReadContext:   resourceZenlayerCloudZecSubnetRead,
		UpdateContext: resourceZenlayerCloudZecSubnetUpdate,
		DeleteContext: resourceZenlayerCloudZecSubnetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.All(
			ipv6TypeChangeForNewFunc(),
			ipv4CidrChangeForNewFunc(),
		),
		Schema: map[string]*schema.Schema{

			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Terraform-Subnet",
				Description: "The name of the subnet, the default value is 'Terraform-Subnet'.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the VPC to be associated.",
			},
			"ip_stack_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Subnet IP stack type. Values: `IPv4`, `IPv6`, `IPv4_IPv6`.",
			},
			"ipv6_type": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"cidr_block", "ipv6_type"},
				ValidateFunc: validation.StringInSlice([]string{"Public", "Private"}, false),
				Description:  "The IPv6 type. Valid values: `Public`, `Private`.",
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"cidr_block", "ipv6_type"},
				ValidateFunc: common2.ValidateCIDRNetworkAddress,
				Description:  "The ipv4 cidr block. A network address block which should be a subnet of the three internal network segments (10.0.0.0/8, 172.16.0.0/12 and 192.168.0.0/16).",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The region that the subnet locates at.",
			},
			"ipv6_cidr_block": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The IPv6 network segment.",
			},
			"is_default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether it is the default subnet.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the subnet.",
			},
		},
	}
}

func ipv4CidrChangeForNewFunc() schema.CustomizeDiffFunc {
	return customdiff.ForceNewIfChange("cidr_block", func(ctx context.Context, old, new, meta interface{}) bool {
		//  v4 -> v6
		if new.(string) == "" {
			return true
		}

		// v6 -> v4
		if old.(string) == "" {
			return true
		}

		return false
	})
}

func ipv6TypeChangeForNewFunc() schema.CustomizeDiffFunc {
	return customdiff.ForceNewIfChange("ipv6_type", func(ctx context.Context, old, new, meta interface{}) bool {
		// from Private / Public ->  none
		return old.(string) != ""
	})
}

func resourceZenlayerCloudZecSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_subnet.delete")()

	subnetId := d.Id()

	ZecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := ZecService.DeleteSubnet(ctx, subnetId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound {
				// vpc doesn't exist
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	// 等待1s
	time.Sleep(time.Second)

	return nil
}

func resourceZenlayerCloudZecSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var _ diag.Diagnostics
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	subnetId := d.Id()

	if d.HasChanges("name", "cidr_block") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {

			err := zecService.ModifySubnet(ctx, subnetId, d.Get("name").(string), d.Get("cidr_block").(string))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	// ipv4 -> ipv4 & ipv6
	if d.HasChange("ipv6_type") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {

			err := zecService.AddSubnetIpv6(ctx, subnetId, d.Get("ipv6_type").(string))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecSubnetRead(ctx, d, meta)
}

func resourceZenlayerCloudZecSubnetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	ZecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zec.NewCreateSubnetRequest()
	request.RegionId = d.Get("region_id").(string)
	request.Name = d.Get("name").(string)
	request.VpcId = d.Get("vpc_id").(string)

	hasIpv4 := false
	hasIpv6 := false

	if v, ok := d.GetOk("cidr_block"); ok {
		request.CidrBlock = v.(string)
		hasIpv4 = true
	}
	if v, ok := d.GetOk("ipv6_type"); ok {
		request.Ipv6Type = v.(string)
		hasIpv6 = true
	}

	if hasIpv4 && hasIpv6 {
		request.StackType = "IPv4_IPv6"
	} else if hasIpv4 && !hasIpv6 {
		request.StackType = "IPv4"
	} else if !hasIpv4 && hasIpv6 {
		request.StackType = "IPv6"
	}

	subnetId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := ZecService.client.WithZecClient().CreateSubnet(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create subnet.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create subnet success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.SubnetId == "" {
			err = fmt.Errorf("subnet id is nil")
			return resource.NonRetryableError(err)
		}
		subnetId = response.Response.SubnetId
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(subnetId)

	return resourceZenlayerCloudZecSubnetRead(ctx, d, meta)
}

func resourceZenlayerCloudZecSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	subnetId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var subnet *zec.SubnetInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		subnet, errRet = zecService.DescribeSubnetById(ctx, subnetId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if subnet == nil {
		d.SetId("")
		tflog.Info(ctx, "subnet not exist", map[string]interface{}{
			"subnetId": subnetId,
		})
		return nil
	}

	// subnet info
	_ = d.Set("region_id", subnet.RegionId)
	_ = d.Set("name", subnet.Name)
	_ = d.Set("cidr_block", subnet.CidrBlock)
	_ = d.Set("vpc_id", subnet.VpcId)
	_ = d.Set("is_default", subnet.IsDefault)
	_ = d.Set("ipv6_cidr_block", subnet.Ipv6CidrBlock)
	_ = d.Set("ip_stack_type", subnet.StackType)
	_ = d.Set("ipv6_type", subnet.Ipv6Type)
	_ = d.Set("create_time", subnet.CreateTime)
	return diags
}
