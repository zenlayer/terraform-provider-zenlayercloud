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
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceZenlayerCloudZecSnapshotCreate,
		ReadContext:   ResourceZenlayerCloudZecSnapshotRead,
		UpdateContext: ResourceZenlayerCloudZecSnapshotUpdate,
		DeleteContext: ResourceZenlayerCloudZecSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"disk_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of disk which the snapshot created from.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Snapshot",
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description:  "The name of the snapshot. The name should start and end with a number or a letter, containing 2 to 63 characters. Only letters, numbers, - and periods (.) are supported.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The availability zone of snapshot.",
			},
			"snapshot_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the snapshot to be queried. Valid values: `Auto`, `Manual`.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the snapshot.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of snapshot. Valid values: `CREATING`, `AVAILABLE`, `FAILED`, `ROLLING_BACK`, `DELETING`.",
			},
			"retention_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Retention time of snapshot.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of resource group grouped snapshot.",
			},
			"resource_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Name of resource group grouped snapshot.",
			},
			"disk_ability": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the snapshot can be used to create a disk.",
			},
		},
	}
}

func ResourceZenlayerCloudZecSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	snapshotId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteSnapshot(ctx, snapshotId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		instance, errRet := zecService.DescribeSnapshotById(ctx, snapshotId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if ok {
				if ee.Code == common2.ResourceNotFound {
					return nil
				}
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if instance == nil {
			return nil
		}

		if *instance.Status == SnapshotDeleting {
			return resource.RetryableError(fmt.Errorf("waiting for load snapshot %s deleting, current status: %s", snapshotId, *instance.Status))
		}

		return resource.NonRetryableError(fmt.Errorf("snapshot status is not deleted, current status %s", *instance.Status))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceZenlayerCloudZecSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	snapId := d.Id()

	if d.HasChange("name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := zec2.NewModifySnapshotsAttributeRequest()
			request.SnapshotName = common.String(d.Get("name").(string))
			request.SnapshotIds = []string{snapId}
			_, err := zecService.client.WithZecClient().ModifySnapshotsAttribute(request)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceZenlayerCloudZecSnapshotRead(ctx, d, meta)
}

func ResourceZenlayerCloudZecSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	vmService := &ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec2.NewCreateSnapshotRequest()
	request.DiskId = common.String(d.Get("disk_id").(string))
	request.SnapshotName = common.String(d.Get("name").(string))

	snapshotId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithZecClient().CreateSnapshot(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create snapshot.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create data snapshot success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.SnapshotId == nil {
			err = fmt.Errorf("snapshot id is nil")
			return resource.NonRetryableError(err)
		}
		snapshotId = *response.Response.SnapshotId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(snapshotId)

	stateConf := BuildSnapshotState(vmService, snapshotId, ctx, d)

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for snapshot (%s) to be created: %v", d.Id(), err))
	}

	return ResourceZenlayerCloudZecSnapshotRead(ctx, d, meta)
}

func ResourceZenlayerCloudZecSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	snapshotId := d.Id()

	vmService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var snapshot *zec2.SnapshotInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		snapshot, errRet = vmService.DescribeSnapshotById(ctx, snapshotId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if snapshot == nil || *snapshot.Status == SnapshotFailed {
		d.SetId("")
		tflog.Info(ctx, "snapshot not exist or created failed", map[string]interface{}{
			"snapshotId": snapshotId,
		})
		return nil
	}

	// snapshot info
	_ = d.Set("disk_id", snapshot.DiskId)
	_ = d.Set("name", snapshot.SnapshotName)
	_ = d.Set("availability_zone", snapshot.ZoneId)
	_ = d.Set("snapshot_type", *snapshot.SnapshotType)
	_ = d.Set("create_time", snapshot.CreateTime)
	_ = d.Set("status", snapshot.Status)
	_ = d.Set("retention_time", snapshot.RetentionTime)
	_ = d.Set("resource_group_id", snapshot.ResourceGroup.ResourceGroupId)
	_ = d.Set("resource_group_name", snapshot.ResourceGroup.ResourceGroupName)
	_ = d.Set("disk_ability", snapshot.DiskAbility)

	return diags

}

func BuildSnapshotState(zecService *ZecService, diskId string, ctx context.Context, d *schema.ResourceData) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{
			SnapshotCreating,
		},
		Target: []string{
			SnapshotAvailable,
		},
		Refresh:        zecService.SnapshotStateRefreshFunc(ctx, diskId, []string{SnapshotFailed}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          3 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}
}
