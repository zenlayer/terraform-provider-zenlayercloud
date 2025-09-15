package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecSnapshotPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecSnapshotPolicyAttachmentCreate,
		ReadContext:   resourceZenlayerCloudZecSnapshotPolicyAttachmentRead,
		DeleteContext: resourceZenlayerCloudZecSnapshotPolicyAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"disk_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the disk. Note: system disk is not support yet.",
			},
			"auto_snapshot_policy_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the auto snapshot policy.",
			},
		},
	}
}

func resourceZenlayerCloudZecSnapshotPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	autoSnapshotPolicyId := d.Get("auto_snapshot_policy_id").(string)
	diskId := d.Get("disk_id").(string)

	request := zec2.NewApplyAutoSnapshotPolicyRequest()
	request.AutoSnapshotPolicyId = common.String(autoSnapshotPolicyId)
	request.DiskIds = []string{diskId}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate)-time.Minute, func() *resource.RetryError {
		_, err := zecService.client.WithZecClient().ApplyAutoSnapshotPolicy(request)
		if err != nil {
			return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(diskId)

	return resourceZenlayerCloudZecSnapshotPolicyAttachmentRead(ctx, d, meta)
}

func resourceZenlayerCloudZecSnapshotPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	diskId := d.Id()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var diskInfo *zec2.DiskInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		diskInfo, errRet = zecService.DescribeDiskById(ctx, diskId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if diskInfo == nil {
		d.SetId("")
		return nil
	}

	if diskInfo.AutoSnapshotPolicyId == nil || *diskInfo.AutoSnapshotPolicyId == "" {
		d.SetId("")
		return nil
	}

	_ = d.Set("auto_snapshot_policy_id", diskInfo.AutoSnapshotPolicyId)
	_ = d.Set("disk_id", diskId)

	return diags
}

func resourceZenlayerCloudZecSnapshotPolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	diskId := d.Id()

	var diskInfo *zec2.DiskInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		diskInfo, errRet = zecService.DescribeDiskById(ctx, diskId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if diskInfo == nil || diskInfo.AutoSnapshotPolicyId == nil || *diskInfo.AutoSnapshotPolicyId == "" {
		return nil
	}

	request := zec2.NewCancelAutoSnapshotPolicyRequest()
	request.AutoSnapshotPolicyId = diskInfo.AutoSnapshotPolicyId
	request.DiskIds = []string{diskId}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, err := zecService.client.WithZecClient().CancelAutoSnapshotPolicy(request)
		if err != nil {
			ee, ok := err.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, err)
			} else if ee.Code == "INVALID_AUTO_SNAPSHOT_POLICY_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
				return nil
			}
			return common2.RetryError(ctx, err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
