/*
Provides a resource to create a ZEC high-availability virtual IP (HaVip).

~> NOTE: Make sure the target subnet has available private IP addresses. If `ip_address` is omitted, the system will allocate one automatically from the subnet; if specified, it must be an available IP within the subnet's CIDR block.

Example Usage

```hcl
resource "zenlayercloud_zec_havip" "example" {
  subnet_id = "subnet-xxxxxxxx"
  name      = "example-havip"
}
```

Import

HaVip can be imported using the id, e.g.

```
terraform import zenlayercloud_zec_havip.example havip-xxxxxxxx
```
*/
package zec

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/zrm"
	sdkcommon "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecHaVip() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecHaVipCreate,
		ReadContext:   resourceZenlayerCloudZecHaVipRead,
		UpdateContext: resourceZenlayerCloudZecHaVipUpdate,
		DeleteContext: resourceZenlayerCloudZecHaVipDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the subnet to which the HaVip belongs.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-HaVip",
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The name of the HaVip. Length must be between 1 and 64 characters. Default is `Terraform-HaVip`.",
			},
			"ip_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsIPv4Address,
				Description:  "The private IPv4 address of the HaVip. Must be one of the available IPs in the subnet's CIDR block. If not specified, the system will allocate one automatically from the subnet.",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The ID of the security group. If not specified, the default security group of the VPC will be used.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The tags associated with the HaVip.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the VPC to which the HaVip belongs.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The region ID where the HaVip is located.",
			},
			"associated_instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The list of instance IDs associated with the HaVip.",
			},
			"master_instance_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the current master instance. Null when no instance is bound.",
			},
			"associated_eips": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of EIPs associated with the HaVip.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"eip_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the EIP.",
						},
						"eip_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The EIP address.",
						},
					},
				},
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the HaVip.",
			},
		},
	}
}

func resourceZenlayerCloudZecHaVipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_havip.create")()

	request := zec.NewCreateHaVipRequest()
	request.SubnetId = sdkcommon.String(d.Get("subnet_id").(string))
	request.Name = sdkcommon.String(d.Get("name").(string))

	if v, ok := d.GetOk("ip_address"); ok {
		request.IpAddress = sdkcommon.String(v.(string))
	}
	if v, ok := d.GetOk("security_group_id"); ok {
		request.SecurityGroupId = sdkcommon.String(v.(string))
	}

	if tags := common.GetTags(d, "tags"); len(tags) > 0 {
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

	haVipId := ""
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithZec2Client().CreateHaVip(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create HaVip.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common.ToJsonString(request),
				"err":     err.Error(),
			})
			return common.RetryError(ctx, err, common.OperationTimeout)
		}

		tflog.Info(ctx, "Create HaVip success.", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common.ToJsonString(request),
			"response": common.ToJsonString(response),
		})

		if response.Response.HaVipId == nil {
			return resource.NonRetryableError(fmt.Errorf("havip id is nil"))
		}
		haVipId = *response.Response.HaVipId
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(haVipId)
	return resourceZenlayerCloudZecHaVipRead(ctx, d, meta)
}

func resourceZenlayerCloudZecHaVipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_havip.read")()

	haVipId := d.Id()
	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}

	var haVip *zec.HaVipInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		haVip, errRet = zecService.DescribeHaVipById(ctx, haVipId)
		if errRet != nil {
			return common.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if haVip == nil {
		d.SetId("")
		return diag.Diagnostics{
			{
				Severity: diag.Warning,
				Summary:  "HaVip not found",
				Detail:   fmt.Sprintf("HaVip %s does not exist", haVipId),
			},
		}
	}

	_ = d.Set("subnet_id", haVip.SubnetId)
	_ = d.Set("name", haVip.Name)
	_ = d.Set("ip_address", haVip.IpAddress)
	_ = d.Set("security_group_id", haVip.SecurityGroupId)
	_ = d.Set("vpc_id", haVip.VpcId)
	_ = d.Set("region_id", haVip.RegionId)
	_ = d.Set("associated_instances", haVip.AssociatedInstances)
	_ = d.Set("master_instance_id", haVip.MasterInstanceId)
	_ = d.Set("create_time", haVip.CreateTime)

	associatedEips := make([]map[string]interface{}, 0, len(haVip.AssociatedEips))
	for _, eip := range haVip.AssociatedEips {
		associatedEips = append(associatedEips, map[string]interface{}{
			"eip_id":      eip.EipId,
			"eip_address": eip.EipAddress,
		})
	}
	_ = d.Set("associated_eips", associatedEips)

	tagMap, err := common.TagsToMap(haVip.Tags)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("tags", tagMap)

	return nil
}

func resourceZenlayerCloudZecHaVipUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_havip.update")()

	haVipId := d.Id()

	if d.HasChange("name") {
		request := zec.NewModifyHaVipAttributeRequest()
		request.HaVipId = sdkcommon.String(haVipId)
		request.Name = sdkcommon.String(d.Get("name").(string))

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			_, err := meta.(*connectivity.ZenlayerCloudClient).WithZec2Client().ModifyHaVipAttribute(request)
			if err != nil {
				return common.RetryError(ctx, err, common.InternalServerError, sdkcommon.NetworkError)
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("tags") {
		zrmService := zrm.NewZrmService(meta.(*connectivity.ZenlayerCloudClient))
		if err := zrmService.ModifyResourceTags(ctx, d, haVipId); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudZecHaVipRead(ctx, d, meta)
}

func resourceZenlayerCloudZecHaVipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_havip.delete")()

	haVipId := d.Id()
	request := zec.NewDeleteHaVipRequest()
	request.HaVipId = sdkcommon.String(haVipId)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, err := meta.(*connectivity.ZenlayerCloudClient).WithZec2Client().DeleteHaVip(request)
		if err != nil {
			sdkErr, ok := err.(*sdkcommon.ZenlayerCloudSdkError)
			if ok && (sdkErr.Code == common.ResourceNotFound || sdkErr.Code == INVALID_HAVIP_NOT_FOUND) {
				return nil
			}
			return common.RetryError(ctx, err, common.InternalServerError, sdkcommon.NetworkError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
