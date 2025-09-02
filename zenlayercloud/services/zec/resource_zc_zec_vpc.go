package zec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudGlobalVpc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudGlobalVpcCreate,
		ReadContext:   resourceZenlayerCloudGlobalVpcRead,
		UpdateContext: resourceZenlayerCloudGlobalVpcUpdate,
		DeleteContext: resourceZenlayerCloudGlobalVpcDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Global-VPC",
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The name of the global VPC.",
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: common2.ValidateCIDRNetworkAddress,
				Description:  "A network address block which should be a subnet of the three internal network segments (10.0.0.0/8, 172.16.0.0/12 and 192.168.0.0/16).",
			},
			"mtu": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1500,
				ValidateFunc: validation.IntInSlice([]int{1300, 1500, 9000}),
				ForceNew:     true,
				Description:  "The maximum transmission unit. This value cannot be changed.",
			},
			"enable_ipv6": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to enable the private IPv6 network segment.",
			},
			"ipv6_cidr_block": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The private IPv6 network segment after `enable_ipv6` is set to `true`.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the global VPC belongs to, default to ID of Default Resource Group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the VPC belongs to, default to Default Resource Group.",
			},
			"is_default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether it is the default VPC.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the global VPC.",
			},
		},
	}
}

func resourceZenlayerCloudGlobalVpcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc.delete")()

	vpcId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteVpc(ctx, vpcId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound || ee.Code == INVALID_VPC_NOT_FOUND {
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

func resourceZenlayerCloudGlobalVpcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc.update")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	vpcId := d.Id()
	d.Partial(true)
	if d.HasChanges("name", "cidr_block", "enable_ipv6") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := zec.NewModifyVpcAttributeRequest()
			request.VpcId = &vpcId
			request.VpcName = common.String(d.Get("name").(string))
			request.CidrBlock = common.String(d.Get("cidr_block").(string))

			if d.HasChange("enable_ipv6") {
				request.EnableIPv6 = common.Bool(d.Get("enable_ipv6").(bool))
			}
			err := zecService.ModifyVpcAttribute(ctx, request)
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
			request.Resources = []*string{common.String(vpcId)}

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

	d.Partial(false)
	return resourceZenlayerCloudGlobalVpcRead(ctx, d, meta)
}

func resourceZenlayerCloudGlobalVpcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := zec.NewCreateVpcRequest()
	request.CidrBlock = d.Get("cidr_block").(string)
	request.Name = d.Get("name").(string)
	request.Mtu = d.Get("mtu").(int)
	request.EnablePriIpv6 = d.Get("enable_ipv6").(bool)

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	vpcId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZecClient().CreateVpc(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create global vpc.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create global vpc success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.VpcId == "" {
			err = fmt.Errorf("vpc id is nil")
			return resource.NonRetryableError(err)
		}
		vpcId = response.Response.VpcId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(vpcId)

	return resourceZenlayerCloudGlobalVpcRead(ctx, d, meta)
}

func resourceZenlayerCloudGlobalVpcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_vpc.read")()

	var diags diag.Diagnostics

	vpcId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var vpc *zec.VpcInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		vpc, errRet = zecService.DescribeVpcById(ctx, vpcId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	// vpc info
	_ = d.Set("name", vpc.Name)
	_ = d.Set("cidr_block", vpc.CidrBlock)
	_ = d.Set("ipv6_cidr_block", vpc.Ipv6CidrBlock)
	_ = d.Set("enable_ipv6", vpc.Ipv6CidrBlock != "")
	_ = d.Set("mtu", vpc.Mtu)
	_ = d.Set("is_default", vpc.IsDefault)
	_ = d.Set("resource_group_id", *vpc.ResourceGroup.ResourceGroupId)
	_ = d.Set("resource_group_name", *vpc.ResourceGroup.ResourceGroupName)
	_ = d.Set("create_time", vpc.CreateTime)

	return diags

}
