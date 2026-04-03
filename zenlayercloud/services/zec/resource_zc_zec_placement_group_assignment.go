package zec

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func ResourceZenlayerCloudZecPlacementGroupAssignment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecPlacementGroupAssignmentCreate,
		ReadContext:   resourceZenlayerCloudZecPlacementGroupAssignmentRead,
		DeleteContext: resourceZenlayerCloudZecPlacementGroupAssignmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The instance ID to assign to the placement group.",
			},
			"placement_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The placement group ID.",
			},
		},
	}
}

func resourceZenlayerCloudZecPlacementGroupAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_placement_group_assignment.create")()

	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)

	instanceId := d.Get("instance_id").(string)
	placementGroupId := d.Get("placement_group_id").(string)

	request := zec.NewModifyInstancePlacementRequest()
	request.InstanceId = common.String(instanceId)
	request.PlacementGroupId = common.String(placementGroupId)

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zenlayerCloudClient.WithZec2Client().ModifyInstancePlacement(request)
		defer common2.LogApiRequest(ctx, "ModifyInstancePlacement", request, response, err)
		if err != nil {
			return common2.RetryError(ctx, err, common2.OperationTimeout)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("fail to assign instance to placement group: %v", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceId, placementGroupId))

	return resourceZenlayerCloudZecPlacementGroupAssignmentRead(ctx, d, meta)
}

func resourceZenlayerCloudZecPlacementGroupAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_placement_group_assignment.read")()

	zecService := &ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}

	instanceId, placementGroupId, err := parsePlacementGroupAssignmentId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var instance *zec.InstanceInfo
	var errRet error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		instance, errRet = zecService.DescribeInstanceById(ctx, instanceId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if instance == nil || common.ToString(instance.PlacementGroupId) != placementGroupId {
		d.SetId("")
		return nil
	}

	_ = d.Set("instance_id", instanceId)
	_ = d.Set("placement_group_id", placementGroupId)

	return nil
}

func resourceZenlayerCloudZecPlacementGroupAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_placement_group_assignment.delete")()

	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)

	instanceId, _, err := parsePlacementGroupAssignmentId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Empty PlacementGroupId means remove from placement group
	request := zec.NewModifyInstancePlacementRequest()
	request.InstanceId = common.String(instanceId)

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		response, errRet := zenlayerCloudClient.WithZec2Client().ModifyInstancePlacement(request)
		defer common2.LogApiRequest(ctx, "ModifyInstancePlacement", request, response, errRet)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet, common2.InternalServerError)
			}
			if ee.Code == common2.ResourceNotFound {
				return nil
			}
			return resource.NonRetryableError(errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("fail to remove instance from placement group: %v", err))
	}

	return nil
}

func parsePlacementGroupAssignmentId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid placement group assignment id format: %s, expected instance_id:placement_group_id", id)
	}
	return parts[0], parts[1], nil
}
