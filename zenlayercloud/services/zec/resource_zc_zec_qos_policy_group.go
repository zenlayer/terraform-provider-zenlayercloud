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
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecQosPolicyGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecQosPolicyGroupCreate,
		ReadContext:   resourceZenlayerCloudZecQosPolicyGroupRead,
		UpdateContext: resourceZenlayerCloudZecQosPolicyGroupUpdate,
		DeleteContext: resourceZenlayerCloudZecQosPolicyGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The region ID where the QoS policy group is located.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The name of the QoS policy group.",
			},
			"bandwidth_limit": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "The shared bandwidth limit in Mbps.",
			},
			"rate_limit_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"LOOSE", "STRICT"}, false),
				Description:  "The rate limit mode of the QoS policy group. Default is LOOSE, Valid values: `LOOSE` - each forwarding server starts with the full group cap, allowing a single connection to reach maximum speed immediately but may briefly exceed the cap under concurrent traffic; `STRICT` - bandwidth is divided evenly across forwarding servers so the group cap is never exceeded, but multiple parallel flows are needed to fully utilize the cap.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group ID the QoS policy group belongs to. Defaults to the default resource group.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the QoS policy group belongs to.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The tags of the QoS policy group.",
			},
			"member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of members currently in the QoS policy group.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the QoS policy group.",
			},
		},
	}
}

func resourceZenlayerCloudZecQosPolicyGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_qos_policy_group.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec.NewCreateQosPolicyGroupRequest()
	request.RegionId = sdkcommon.String(d.Get("region_id").(string))
	request.Name = sdkcommon.String(d.Get("name").(string))
	request.BandwidthLimit = sdkcommon.Int64(int64(d.Get("bandwidth_limit").(int)))
	request.RateLimitMode = sdkcommon.String(d.Get("rate_limit_mode").(string))

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = sdkcommon.String(v.(string))
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

	groupId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZec2Client().CreateQosPolicyGroup(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create QoS policy group.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common.ToJsonString(request),
				"err":     err.Error(),
			})
			return common.RetryError(ctx, err, common.OperationTimeout)
		}

		if response.Response.QosPolicyGroupId == nil {
			return resource.NonRetryableError(fmt.Errorf("qos policy group id is nil"))
		}
		groupId = *response.Response.QosPolicyGroupId
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(groupId)

	return resourceZenlayerCloudZecQosPolicyGroupRead(ctx, d, meta)
}

func resourceZenlayerCloudZecQosPolicyGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_qos_policy_group.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	groupId := d.Id()

	var group *zec.QosPolicyGroup
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var errRet error
		group, errRet = zecService.DescribeQosPolicyGroupById(ctx, groupId)
		if errRet != nil {
			return common.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if group == nil {
		d.SetId("")
		tflog.Info(ctx, "QoS policy group does not exist", map[string]interface{}{
			"groupId": groupId,
		})
		return nil
	}

	_ = d.Set("region_id", group.RegionId)
	_ = d.Set("name", group.Name)
	_ = d.Set("bandwidth_limit", int(sdkcommon.ToInt64(group.BandwidthLimit)))
	_ = d.Set("rate_limit_mode", group.RateLimitMode)
	_ = d.Set("member_count", group.MemberCount)
	_ = d.Set("create_time", group.CreateTime)

	if group.ResourceGroup != nil {
		_ = d.Set("resource_group_id", group.ResourceGroup.ResourceGroupId)
		_ = d.Set("resource_group_name", group.ResourceGroup.ResourceGroupName)
	}

	tagMap, errRet := common.TagsToMap(group.Tags)
	if errRet != nil {
		return diag.FromErr(errRet)
	}
	_ = d.Set("tags", tagMap)

	return nil
}

func resourceZenlayerCloudZecQosPolicyGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_qos_policy_group.update")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	groupId := d.Id()
	d.Partial(true)

	if d.HasChanges("name", "bandwidth_limit", "rate_limit_mode") {
		request := zec.NewModifyQosPolicyGroupRequest()
		request.QosPolicyGroupId = sdkcommon.String(groupId)

		if d.HasChange("name") {
			request.Name = sdkcommon.String(d.Get("name").(string))
		}
		if d.HasChange("bandwidth_limit") {
			request.BandwidthLimit = sdkcommon.Int64(int64(d.Get("bandwidth_limit").(int)))
		}
		if d.HasChange("rate_limit_mode") {
			request.RateLimitMode = sdkcommon.String(d.Get("rate_limit_mode").(string))
		}

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			_, err := zecService.client.WithZec2Client().ModifyQosPolicyGroup(request)
			if err != nil {
				return common.RetryError(ctx, err, common.InternalServerError)
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
			request.ResourceGroupId = sdkcommon.String(d.Get("resource_group_id").(string))
			request.Resources = []string{groupId}

			_, err := meta.(*connectivity.ZenlayerCloudClient).WithUsrClient().AddResourceResourceGroup(request)
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
		err := zrmService.ModifyResourceTags(ctx, d, groupId)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)
	return resourceZenlayerCloudZecQosPolicyGroupRead(ctx, d, meta)
}

func resourceZenlayerCloudZecQosPolicyGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "resource.zenlayercloud_zec_qos_policy_group.delete")()

	groupId := d.Id()

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		request := zec.NewDeleteQosPolicyGroupRequest()
		request.QosPolicyGroupId = sdkcommon.String(groupId)

		_, err := meta.(*connectivity.ZenlayerCloudClient).WithZec2Client().DeleteQosPolicyGroup(request)
		if err != nil {
			if sdkErr, ok := err.(*sdkcommon.ZenlayerCloudSdkError); ok {
				if sdkErr.Code == common.ResourceNotFound {
					return nil
				}
			}
			return common.RetryError(ctx, err, common.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
