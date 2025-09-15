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
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func ResourceZenlayerCloudZecDisk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecDiskCreate,
		ReadContext:   resourceZenlayerCloudVmDiskRead,
		UpdateContext: resourceZenlayerCloudVmDiskUpdate,
		DeleteContext: resourceZenlayerCloudVmDiskDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of zone that the disk locates at.",
			},
			"disk_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Disk",
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The name of the disk.",
			},
			"disk_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(20),
				Description:  "The size of disk. Unit: GiB. The minimum value is 20 GiB. When resize the disk, the new size must be greater than the former value.",
			},
			"disk_category": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "Standard NVMe SSD",
				Description: "The category of disk.",
			},
			"disk_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the disk. Values are: `SYSTEM`, `DATA`.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource group id the disk belongs to.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicate whether to force delete the data disk. Default is `false`. If set true, the disk will be permanently deleted instead of being moved into the recycle bin.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the disk.",
			},
		},
	}
}

func resourceZenlayerCloudVmDiskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diskId := d.Id()

	// force_delete: terminate and then delete
	forceDelete := d.Get("force_delete").(bool)

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteDiskById(ctx, diskId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	notExist := false

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		disk, errRet := zecService.DescribeDiskById(ctx, diskId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if disk == nil {
			notExist = true
			return nil
		}

		if disk.DiskStatus == ZecDiskStatusRecycle {
			//in recycling
			return nil
		}

		if diskIsOperating(disk.DiskStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for disk %s recycling, current status: %s", disk.DiskId, disk.DiskStatus))
		}

		return resource.NonRetryableError(fmt.Errorf("disk status is not recycle, current status %s", disk.DiskStatus))
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if notExist || !forceDelete {
		return nil
	}
	tflog.Debug(ctx, "Releasing disk ...", map[string]interface{}{
		"diskId": diskId,
	})

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteDiskById(ctx, diskId)
		if errRet != nil {

			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			}
			if ee.Code == INVALID_DISK_NOT_FOUND || ee.Code == common2.ResourceNotFound {
				// disk doesn't exist
				return nil
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		return nil
	})

	if err != nil {
		diag.FromErr(err)
	}
	return nil
}

func resourceZenlayerCloudVmDiskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	diskId := d.Id()
	if d.HasChange("disk_name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := zec2.NewModifyDisksAttributesRequest()
			request.DiskName = d.Get("disk_name").(string)
			request.DiskIds = []string{diskId}
			_, err := zecService.client.WithZecClient().ModifyDisksAttributes(request)
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
			request.Resources = []*string{common.String(diskId)}

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

	d.Partial(true)
	if d.HasChange("disk_size") {

		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := zecService.ResizeDisk(ctx, diskId, d.Get("disk_size").(int))
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudVmDiskRead(ctx, d, meta)
}

func resourceZenlayerCloudZecDiskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	vmService := &ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := zec2.NewCreateDisksRequest()
	request.DiskName = d.Get("disk_name").(string)
	request.DiskSize = d.Get("disk_size").(int)
	request.DiskCategory = d.Get("disk_category").(string)
	request.ZoneId = d.Get("availability_zone").(string)

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	diskId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithZecClient().CreateDisks(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create data disk.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create data disk success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if len(response.Response.DiskIds) < 1 {
			err = fmt.Errorf("disk id is nil")
			return resource.NonRetryableError(err)
		}
		diskId = response.Response.DiskIds[0]

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(diskId)

	stateConf := BuildDiskState(vmService, diskId, ctx, d)

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for disk (%s) to be created: %v", d.Id(), err))
	}

	return resourceZenlayerCloudVmDiskRead(ctx, d, meta)
}

func resourceZenlayerCloudVmDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	diskId := d.Id()

	vmService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var disk *zec2.DiskInfo
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		disk, errRet = vmService.DescribeDiskById(ctx, diskId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}

		if disk != nil && diskIsOperating(disk.DiskStatus) {
			return resource.RetryableError(fmt.Errorf("waiting for disk %s operation", disk.DiskId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if disk == nil || disk.DiskStatus == ZecDiskStatusRecycle {
		d.SetId("")
		tflog.Info(ctx, "disk not exist or is been recycled", map[string]interface{}{
			"diskId": diskId,
		})
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The disk is not exist",
			Detail:   fmt.Sprintf("The disk %s is not exist", diskId),
		})
		return diags
	}

	if disk.DiskStatus == ZecDiskStatusRecycle || disk.DiskStatus == ZecDiskStatusFaileld {
		d.SetId("")
		tflog.Info(ctx, "disk not exist or is been recycled", map[string]interface{}{
			"diskId": diskId,
		})
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The status of disk is invalid",
			Detail:   fmt.Sprintf("The status of disk %s is %s", diskId, disk.DiskStatus),
		})
		return diags
	}
	// disk info
	_ = d.Set("availability_zone", disk.ZoneId)
	_ = d.Set("disk_name", disk.DiskName)
	_ = d.Set("disk_category", disk.DiskCategory)
	_ = d.Set("disk_type", disk.DiskType)
	_ = d.Set("disk_size", disk.DiskSize)
	_ = d.Set("create_time", disk.CreateTime)
	_ = d.Set("resource_group_id", disk.ResourceGroupId)
	_ = d.Set("resource_group_name", disk.ResourceGroupName)

	return diags

}

func BuildDiskState(zecService *ZecService, diskId string, ctx context.Context, d *schema.ResourceData) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{
			ZecDiskStatusAttaching,
			ZecDiskStatusDetaching,
			ZecDiskStatusCreating,
			ZecDiskStatusDeleting,
		},
		Target: []string{
			ZecDiskStatusInUse,
			ZecDiskStatusAvailable,
		},
		Refresh:        zecService.DiskStateRefreshFunc(ctx, diskId, []string{ZecDiskStatusRecycle}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          3 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}
}

func diskIsOperating(status string) bool {
	return common2.IsContains([]string{
		ZecDiskStatusRecycling,
		ZecDiskStatusAttaching,
		ZecDiskStatusDetaching,
		ZecDiskStatusCreating,
		ZecDiskStatusDeleting,
		ZecDiskStatusResizing}, status)
}
