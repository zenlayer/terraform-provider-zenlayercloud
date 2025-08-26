package zec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecDiskAttachment() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecDiskAttachmentCreate,
		ReadContext:   resourceZenlayerCloudZecDiskAttachmentRead,
		DeleteContext: resourceZenlayerCloudZecDiskAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"disk_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the Disk.",
			},
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of a ZEC instance.",
			},
		},
	}
}

func resourceZenlayerCloudZecDiskAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_disk_attachment.delete")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	attachment, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	request := zec2.NewDetachDisksRequest()
	diskId := attachment[0]
	request.DiskIds = []string{diskId}
	request.InstanceCheckFlag = common.Bool(false)

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := zecService.client.WithZecClient().DetachDisks(request)

		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		disk, errRet := zecService.DescribeDiskById(ctx, diskId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		if disk == nil {
			d.SetId("")
		}

		if disk.DiskStatus == ZecDiskStatusAvailable {
			return nil
		}

		if disk != nil && diskIsOperating(disk.DiskStatus) {
			return resource.RetryableError(fmt.Errorf("waiting disk %s operation", disk.DiskId))
		}
		return nil
	})
	return nil
}

func resourceZenlayerCloudZecDiskAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_disk_attachment.create")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	diskId := d.Get("disk_id").(string)
	instanceId := d.Get("instance_id").(string)

	request := zec2.NewAttachDisksRequest()
	request.InstanceId = instanceId
	request.DiskIds = []string{diskId}

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := zecService.client.WithZecClient().AttachDisks(request)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", diskId, instanceId))

	return resourceZenlayerCloudZecDiskAttachmentRead(ctx, d, meta)
}

func resourceZenlayerCloudZecDiskAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zec_disk_attachment.read")()

	vNicInstanceId, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}
	diskId := vNicInstanceId[0]

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var disk *zec2.DiskInfo
	var errRet error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		disk, errRet = zecService.DescribeDiskById(ctx, diskId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		if disk == nil {
			d.SetId("")
		}
		if disk != nil && diskIsOperating(disk.DiskStatus) {
			return resource.RetryableError(fmt.Errorf("waiting disk %s operation", disk.DiskId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if disk == nil || disk.InstanceId == "" {
		d.SetId("")
		return nil
	}

	_ = d.Set("disk_id", disk.DiskId)
	_ = d.Set("instance_id", disk.InstanceId)
	return nil
}
