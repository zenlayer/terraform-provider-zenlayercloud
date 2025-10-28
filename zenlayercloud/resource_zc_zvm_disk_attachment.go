/*
Provide a resource to attach a disk to an instance.

Example Usage

```hcl

resource "zenlayercloud_zvm_disk_attachment" "foo" {
  disk_id 	 	= "diskxxxx"
  instance_id  	= "instancexxxx"
}
```

Import

Disk attachment can be imported, e.g.

```
$ terraform import zenlayercloud_zvm_disk_attachment.foo disk-id:instance-id
```
*/
package zenlayercloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"time"
)

func resourceZenlayerCloudVmDiskAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudVmDiskAttachmentCreate,
		ReadContext:   resourceZenlayerCloudVmDiskAttachmentRead,
		DeleteContext: resourceZenlayerCloudVmDiskAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of instance.",
			},
			"disk_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of disk.",
			},
		},
	}
}

func resourceZenlayerCloudVmDiskAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vmService := &VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	attachment, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	request := vm.NewDetachDisksRequest()
	diskId := attachment[0]
	request.DiskIds = []string{diskId}

	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		_, errRet := vmService.client.WithVmClient().DetachDisks(request)
		ee, ok := errRet.(*common.ZenlayerCloudSdkError)
		if ok {
			if ee.Code == "UNSUPPORTED_OPERATION_DISK_NO_ATTACH" {
				return nil
			}
		}
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	stateConf := BuildDiskState(vmService, diskId, ctx, d)

	disk, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for detachment (%s) to be deleted: %v", d.Id(), err)
	}
	if disk == nil {
		return diag.Errorf("detach disk (%s) failed as disk not found", diskId)
	}
	if disk.(*vm.DiskInfo).DiskStatus != VmDiskStatusAvailable {
		return diag.Errorf("detach disk (%s) failed, current status is %s", diskId, disk.(*vm.DiskInfo).DiskStatus)
	}
	return nil
}

func resourceZenlayerCloudVmDiskAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vmService := &VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	diskId := d.Get("disk_id").(string)
	instanceId := d.Get("instance_id").(string)
	var disk *vm.DiskInfo
	var errRet error

	err := resource.RetryContext(ctx, common2.ReadRetryTimeout, func() *resource.RetryError {
		disk, errRet = vmService.DescribeDiskById(ctx, diskId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if disk == nil {
			return resource.NonRetryableError(fmt.Errorf("disk (%s) is not found", diskId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, common2.ReadRetryTimeout, func() *resource.RetryError {
		instance, errRet := vmService.DescribeInstanceById(ctx, instanceId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if instance == nil {
			return resource.NonRetryableError(fmt.Errorf("instance (%s) is not found", diskId))
		}
		if instance.InstanceStatus == VmInstanceStatusDeloying || instance.InstanceStatus == VmDiskStatusRecycle {
			return resource.NonRetryableError(fmt.Errorf("error instance (%s) status : %s", diskId, instance.InstanceStatus))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if disk.InstanceId != instanceId {
		if disk.DiskStatus != VmDiskStatusAvailable {
			return diag.FromErr(fmt.Errorf("disk (%s) status is illegal %s", diskId, disk.DiskStatus))
		}

		request := vm.NewAttachDisksRequest()
		request.DiskIds = []string{diskId}
		request.InstanceId = instanceId

		if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
			_, errRet := vmService.client.WithVmClient().AttachDisks(request)
			if errRet != nil {
				return common2.RetryError(ctx, errRet)
			}
			return nil
		}); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(diskId + ":" + instanceId)

	stateConf := BuildDiskState(vmService, diskId, ctx, d)

	diskState, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for attachment (%s) to be created: %v", d.Id(), err)
	}
	if diskState == nil {
		return diag.Errorf("attach disk (%s)  to instance failed as  not found", diskId)
	}

	if diskState.(*vm.DiskInfo).DiskStatus == VmDiskStatusAvailable {
		return diag.Errorf("attach disk (%s) to instance (%s) failed", diskId, instanceId)
	}

	return resourceZenlayerCloudVmDiskAttachmentRead(ctx, d, meta)
}

func resourceZenlayerCloudVmDiskAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	vmService := &VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	association, err := common2.ParseResourceId(d.Id(), 2)
	if err != nil {
		return diag.FromErr(err)
	}

	var disk *vm.DiskInfo
	var errRet error

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		disk, errRet = vmService.DescribeDiskById(ctx, association[0])
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

	_ = d.Set("disk_id", association[0])
	_ = d.Set("instance_id", association[1])

	return diags
}
