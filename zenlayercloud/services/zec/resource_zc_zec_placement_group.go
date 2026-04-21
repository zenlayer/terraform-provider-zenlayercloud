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
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/services/zrm"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	user "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/user20240529"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecPlacementGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecPlacementGroupCreate,
		ReadContext:   resourceZenlayerCloudZecPlacementGroupRead,
		UpdateContext: resourceZenlayerCloudZecPlacementGroupUpdate,
		DeleteContext: resourceZenlayerCloudZecPlacementGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The zone ID of the placement group. such as 'asia-east-1a'.",
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description:  "The name of the placement group. Must be 2-63 characters, starting and ending with a letter or digit.",
			},
			"partition_num": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntAtLeast(2),
				Description:  "The number of partitions. Range: 2-5, default: 3.",
			},
			"affinity": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Computed:     true,
				Description:  "The affinity level. Range: 1 to partition_num/2. Default: partition_num/2.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group ID the placement group belongs to.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource group name the placement group belongs to.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The tags of the placement group.",
			},
			"instance_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of instances in the placement group.",
			},
			"instance_ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of instance IDs associated with the placement group.",
			},
			"constraint_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The constraint satisfaction status of the placement group.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the placement group.",
			},
		},
	}
}

func resourceZenlayerCloudZecPlacementGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_placement_group.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec.NewCreatePlacementGroupRequest()
	request.ZoneId = common.String(d.Get("zone_id").(string))
	request.Name = common.String(d.Get("name").(string))

	if v, ok := d.GetOk("partition_num"); ok {
		request.PartitionNum = common.Integer(v.(int))
	}
	if v, ok := d.GetOk("affinity"); ok {
		request.Affinity = common.Integer(v.(int))
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

	placementGroupId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZec2Client().CreatePlacementGroup(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create placement group.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err, common2.OperationTimeout)
		}

		tflog.Info(ctx, "Create placement group success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.PlacementGroupId == nil {
			err = fmt.Errorf("placement group id is nil")
			return resource.NonRetryableError(err)
		}
		placementGroupId = *response.Response.PlacementGroupId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(placementGroupId)

	return resourceZenlayerCloudZecPlacementGroupRead(ctx, d, meta)
}

func resourceZenlayerCloudZecPlacementGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_placement_group.read")()

	var diags diag.Diagnostics

	placementGroupId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var pg *zec.PlacementGroupInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		pg, errRet = zecService.DescribePlacementGroupById(ctx, placementGroupId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	if pg == nil {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The placement group does not exist",
			Detail:   fmt.Sprintf("The placement group %s does not exist", placementGroupId),
		})
		return diags
	}

	_ = d.Set("zone_id", pg.ZoneId)
	_ = d.Set("name", pg.Name)
	_ = d.Set("partition_num", pg.PartitionNum)
	_ = d.Set("affinity", pg.Affinity)
	_ = d.Set("instance_count", pg.InstanceCount)
	_ = d.Set("instance_ids", pg.InstanceIds)
	_ = d.Set("constraint_status", pg.ConstraintStatus)
	_ = d.Set("create_time", pg.CreateTime)

	if pg.ResourceGroup != nil {
		_ = d.Set("resource_group_id", pg.ResourceGroup.ResourceGroupId)
		_ = d.Set("resource_group_name", pg.ResourceGroup.ResourceGroupName)
	}

	tagMap, errRet := common2.TagsToMap(pg.Tags)
	if errRet != nil {
		return diag.FromErr(errRet)
	}
	_ = d.Set("tags", tagMap)

	return diags
}

func resourceZenlayerCloudZecPlacementGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_placement_group.update")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	placementGroupId := d.Id()

	d.Partial(true)

	if d.HasChanges("name", "partition_num", "affinity") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := zec.NewModifyPlacementGroupAttributesRequest()
			request.PlacementGroupId = &placementGroupId

			if d.HasChange("name") {
				request.Name = common.String(d.Get("name").(string))
			}
			if d.HasChange("partition_num") {
				request.PartitionNum = common.Integer(d.Get("partition_num").(int))
			}
			if d.HasChange("affinity") {
				request.Affinity = common.Integer(d.Get("affinity").(int))
			}

			response, err := zecService.client.WithZec2Client().ModifyPlacementGroupAttributes(request)
			defer common2.LogApiRequest(ctx, "ModifyPlacementGroupAttributes", request, response, err)
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
			request.Resources = []string{placementGroupId}

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
		err := zrmService.ModifyResourceTags(ctx, d, placementGroupId)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)
	return resourceZenlayerCloudZecPlacementGroupRead(ctx, d, meta)
}

func resourceZenlayerCloudZecPlacementGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_placement_group.delete")()

	placementGroupId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		request := zec.NewDeletePlacementGroupsRequest()
		request.PlacementGroupIds = []string{placementGroupId}

		response, err := zecService.client.WithZec2Client().DeletePlacementGroups(request)
		defer common2.LogApiRequest(ctx, "DeletePlacementGroups", request, response, err)
		if err != nil {
			ee, ok := err.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, err, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound || ee.Code == INVALID_PLACEMENT_GROUP_NOT_FOUND {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		if response.Response.FailedPlacementGroupIds != nil && len(response.Response.FailedPlacementGroupIds) > 0 {
			return resource.NonRetryableError(fmt.Errorf("failed to delete placement group %s", placementGroupId))
		}

		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
