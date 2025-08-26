/*
Provide a resource to create data disk.

Example Usage

```hcl

resource "zenlayercloud_disk" "foo" {
  availability_zone 	 = "SEL-A"
  name  				 = "SEL-20G"
  disk_size				 = 20
}
```

Import

Disk instance can be imported, e.g.

```
$ terraform import zenlayercloud_disk.test disk-id
```
*/
package zenlayercloud

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
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"time"
)

func resourceZenlayerCloudVmDisk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudVmDiskCreate,
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
			"instance_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of instance which the disk attached to.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Terraform-Disk",
				ValidateFunc: validation.StringLenBetween(1, 64),
				Description:  "The name of the disk.",
			},
			"disk_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(20),
				Description:  "The size of disk. Unit: GB. The minimum value is 20 GB.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The resource group id the disk belongs to.",
			},
			"charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "POSTPAID",
				ValidateFunc: validation.StringInSlice([]string{"POSTPAID", "PREPAID"}, false),
				ForceNew:     true,
				Description:  "Charge type of disk.",
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicate whether to force delete the data disk. Default is `false`. If set true, the disk will be permanently deleted instead of being moved into the recycle bin.",
			},
			"charge_prepaid_period": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "The tenancy (time unit is month) of the prepaid disk.",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Create time of the disk.",
			},
			"expired_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expire time of the disk.",
			},
		},
	}
}

func resourceZenlayerCloudVmDiskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	diskId := d.Id()

	// force_delete: terminate and then delete
	forceDelete := d.Get("force_delete").(bool)

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := vmService.DeleteDisk(ctx, diskId)
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
		disk, errRet := vmService.DescribeDiskById(ctx, diskId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if disk == nil {
			notExist = true
			return nil
		}

		if disk.DiskStatus == VmDiskStatusRecycle {
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
		errRet := vmService.ReleaseDisk(ctx, diskId)
		if errRet != nil {

			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if !ok {
				return common2.RetryError(ctx, errRet)
			}
			if ee.Code == "INVALID_DISK_NOT_FOUND" || ee.Code == common2.ResourceNotFound {
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

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	diskId := d.Id()
	if d.HasChange("name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			err := vmService.ModifyDiskName(ctx, diskId, d.Get("name").(string))
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
			err := vmService.ModifyDiskResourceGroupId(ctx, diskId, d.Get("resource_group_id").(string))
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

func resourceZenlayerCloudVmDiskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	vmService := &VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := vm.NewCreateDisksRequest()
	request.DiskName = d.Get("name").(string)
	request.ChargeType = d.Get("charge_type").(string)
	request.DiskSize = common.Integer(d.Get("disk_size").(int))

	if v, ok := d.GetOk("availability_zone"); ok {
		request.ZoneId = v.(string)
	}

	if v, ok := d.GetOk("instance_id"); ok {
		request.InstanceId = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	if request.ChargeType == "PREPAID" {
		request.ChargePrepaid = &vm.ChargePrepaid{}

		if period, ok := d.GetOk("charge_prepaid_period"); ok {
			request.ChargePrepaid.Period = period.(int)
		} else {
			diags = append(diags, diag.Diagnostic{
				Summary: "Missing required argument",
				Detail:  "charge_prepaid_period is missing on prepaid disk.",
			})
			return diags
		}
	}

	diskId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithVmClient().CreateDisks(request)
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

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var disk *vm.DiskInfo
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

	if disk == nil || disk.DiskStatus == VmDiskStatusRecycle {
		d.SetId("")
		tflog.Info(ctx, "disk not exist or is been recycled", map[string]interface{}{
			"diskId": diskId,
		})
		return nil
	}

	// disk info
	_ = d.Set("availability_zone", disk.ZoneId)
	_ = d.Set("instance_id", disk.InstanceId)
	_ = d.Set("name", disk.DiskName)
	_ = d.Set("disk_size", disk.DiskSize)
	_ = d.Set("charge_type", disk.ChargeType)
	_ = d.Set("create_time", disk.CreateTime)
	_ = d.Set("expired_time", disk.ExpiredTime)
	_ = d.Set("charge_prepaid_period", disk.Period)

	return diags

}

func BuildDiskState(vmService *VmService, diskId string, ctx context.Context, d *schema.ResourceData) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{
			VmDiskStatusAttaching,
			VmDiskStatusDetaching,
			VmDiskStatusCreating,
			VmDiskStatusDeleting,
		},
		Target: []string{
			VmDiskStatusInUse,
			VmDiskStatusAvailable,
		},
		Refresh:        vmService.DiskStateRefreshFunc(ctx, diskId, []string{VmDiskStatusRecycle}),
		Timeout:        d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:          3 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}
}

func diskIsOperating(status string) bool {
	return common2.IsContains([]string{
		VmDiskStatusRecycling,
		VmDiskStatusAttaching,
		VmDiskStatusDetaching,
		VmDiskStatusCreating,
		VmDiskStatusDeleting}, status)
}
