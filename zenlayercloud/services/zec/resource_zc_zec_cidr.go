package zec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecCidr() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecCidrCreate,
		ReadContext:   resourceZenlayerCloudZecCidrRead,
		UpdateContext: resourceZenlayerCloudZecCidrUpdate,
		DeleteContext: resourceZenlayerCloudZecCidrDelete,
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
				Description: "The region ID that the public CIDR block locates at.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description: "Name of the public CIDR block.The name should start and end with a number or a letter, containing 2 to 63 characters. Only letters, numbers, -, slash(/) and periods (.) are supported.",
			},
			"network_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringInSlice([]string{"BGPLine", "CN2Line", "LocalLine", "ChinaTelecom", "ChinaUnicom", "ChinaMobile", "Cogent"}, false),
				Description:   "Network types of public CIDR block. Valid values: `BGPLine`, `CN2Line`, `LocalLine`, `ChinaTelecom`, `ChinaUnicom`, `ChinaMobile`, `Cogent`.",
			},
			"netmask": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(27, 30),
				Description:  "Netmask of CIDR block. Valid values: `27` to `30`.",
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
			"cidr_block_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public CIDR block address.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the public CIDR block.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the public CIDR block.",
			},
		},
	}
}

func resourceZenlayerCloudZecCidrCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_cidr.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec.NewCreateCidrRequest()
	request.RegionId = common2.String(d.Get("region_id").(string))
	request.EipV4Type = common2.String(d.Get("network_type").(string))
	request.Netmask = &zec.NetmaskInfo{
		Netmask: common2.Integer(d.Get("netmask").(int)),
		Amount:  common2.Integer(1),
	}
	if v, ok := d.GetOk("name"); ok {
		request.Name = common2.String(v.(string))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common2.String(v.(string))
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZec2Client().CreateCidr(request)
		if err != nil {
			return common.RetryError(ctx, err)
		}

		if len(response.Response.CidrIds) == 0 {
			return resource.NonRetryableError(fmt.Errorf("failed to get CIDR block ID from response"))
		}

		d.SetId(response.Response.CidrIds[0])
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceZenlayerCloudZecCidrRead(ctx, d, meta)
}

func resourceZenlayerCloudZecCidrRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_cidr.read")()

	var diags diag.Diagnostics
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	cidrId := d.Id()
	// 等待创建完成

	var cidrInfo *zec.CidrInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		cidrBlock, errRet := zecService.DescribeCidrById(ctx, cidrId)
		if errRet != nil {
			return common.RetryError(ctx, errRet)
		}

		if cidrBlock != nil && cidrIsOperating(*cidrBlock.Status) {
			return resource.RetryableError(fmt.Errorf("waiting for cidrInfo %s operation", cidrBlock.CidrId))
		}
		cidrInfo = cidrBlock
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if cidrInfo == nil {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The cidr is not exist",
			Detail:   fmt.Sprintf("The cidr block %s is not exist", cidrId),
		})
		return diags
	}

	if *cidrInfo.Status == CidrStatusFailed ||
		*cidrInfo.Status == CidrStatusRecycled {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The cidr is failed or recycled",
			Detail:   fmt.Sprintf("The status of cidr block %s is %s", cidrId, *cidrInfo.Status),
		})
		return diags
	}

	_ = d.Set("region_id", cidrInfo.RegionId)
	_ = d.Set("name", cidrInfo.Name)
	_ = d.Set("network_type", cidrInfo.EipV4Type)
	_ = d.Set("cidr_block_address", cidrInfo.CidrBlock)
	_ = d.Set("resource_group_id", cidrInfo.ResourceGroupId)
	_ = d.Set("resource_group_name", cidrInfo.ResourceGroupName)
	_ = d.Set("status", cidrInfo.Status)
	_ = d.Set("netmask", cidrInfo.Netmask)
	_ = d.Set("create_time", cidrInfo.CreateTime)

	return nil
}

func cidrIsOperating(status string) bool {
	return common.IsContains([]string{CidrStatusCreating, CidrStatusRecycling, CidrStatusDeleting}, status)
}

func resourceZenlayerCloudZecCidrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_cidr.update")()
	//
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	d.Partial(true)
	cidrId := d.Id()

	if d.HasChange("name") {
		request := zec.NewModifyCidrAttributeRequest()
		request.CidrId = common2.String(cidrId)
		request.Name = common2.String(d.Get("name").(string))
		_, err := zecService.client.WithZec2Client().ModifyCidrAttribute(request)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("resource_group_id") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := user.NewAddResourceResourceGroupRequest()
			request.ResourceGroupId = common2.String(d.Get("resource_group_id").(string))
			request.Resources = []*string{common2.String(cidrId)}

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

	return resourceZenlayerCloudZecCidrRead(ctx, d, meta)
}

func resourceZenlayerCloudZecCidrDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_cidr.delete")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec.NewDeleteCidrRequest()
	cidrId := d.Id()
	request.CidrId = common2.String(cidrId)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := zecService.client.WithZec2Client().DeleteCidr(request)
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
		cidrInfo, errRet := zecService.DescribeCidrById(ctx, cidrId)
		if errRet != nil {
			return common.RetryError(ctx, errRet, common.InternalServerError)
		}
		if cidrInfo == nil {
			notExist = true
			return nil
		}

		if *cidrInfo.Status == CidrStatusRecycled {
			return nil
		}
		if *cidrInfo.Status == CidrStatusRecycling {
			return resource.RetryableError(fmt.Errorf("zec cidrInfo (%s) is recycling", cidrId))
		}
		if *cidrInfo.Status == CidrStatusDeleting {
			return resource.RetryableError(fmt.Errorf("zec cidrInfo (%s) is deleting", cidrId))
		}

		return resource.NonRetryableError(fmt.Errorf("zec cidr status is not recycle, current status %s", *cidrInfo.Status))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if notExist {
		return nil
	}

	tflog.Debug(ctx, "Releasing zec CIDR block ...", map[string]interface{}{
		"cidrId": cidrId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, err := zecService.client.WithZec2Client().DeleteCidr(request)
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

	return resourceZenlayerCloudZecCidrRead(ctx, d, meta)
}
