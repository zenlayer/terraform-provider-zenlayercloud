/*
Provides a resource to manage image.

~> **NOTE:** You have to keep the instance power off if the image is created from instance.

Example Usage

```hcl
resource "zenlayercloud_zvm_image" "foo" {
  image_name       	= "web-image-centos"
  instance_id    	= "xxxxxx"
  image_description	= "create a image by the web server"
}
```

Import

Image can be imported, e.g.

```
$ terraform import zenlayercloud_zvm_image.foo img-xxxxxxx
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
	"strconv"
	"time"
)

func resourceZenlayerCloudVmImage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudImageCreate,
		ReadContext:   resourceZenlayerCloudImageRead,
		UpdateContext: resourceZenlayerCloudImageUpdate,
		DeleteContext: resourceZenlayerCloudImageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "VM instance ID.",
			},
			"image_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(2, 24),
				Description:  "Image name. Cannot be modified unless recreated.",
			},
			"image_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(2, 256),
				Description:  "Image description.",
			},
			"image_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Image size.",
			},
		},
	}
}

func resourceZenlayerCloudImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zvm_image.delete")()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	imageId := d.Id()

	// delete image
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := vmService.DeleteImage(ctx, imageId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceZenlayerCloudImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	imageId := d.Id()
	if d.HasChanges("image_name", "image_description") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			imageName := d.Get("image_name").(string)
			imageDesc := d.Get("image_description").(string)
			err := vmService.ModifyImage(ctx, imageId, imageName, imageDesc)
			if err != nil {
				return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
			}
			return nil
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceZenlayerCloudImageRead(ctx, d, meta)
}

func resourceZenlayerCloudImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayercloud_zvm_image.create")()

	request := vm.NewCreateImageRequest()
	request.ImageName = d.Get("image_name").(string)
	request.InstanceId = d.Get("instance_id").(string)
	if v, ok := d.GetOk("image_description"); ok {
		request.ImageDescription = v.(string)
	}
	imageId := ""

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := meta.(*connectivity.ZenlayerCloudClient).WithVmClient().CreateImages(request)
		if err != nil {
			tflog.Info(ctx, "Fail to create image.", map[string]interface{}{
				"action":  request.GetAction(),
				"request": common2.ToJsonString(request),
				"err":     err.Error(),
			})
			return common2.RetryError(ctx, err)
		}

		tflog.Info(ctx, "Create image success", map[string]interface{}{
			"action":   request.GetAction(),
			"request":  common2.ToJsonString(request),
			"response": common2.ToJsonString(response),
		})

		if response.Response.ImageId == "" {
			err = fmt.Errorf("image id is nil")
			return resource.NonRetryableError(err)
		}
		imageId = response.Response.ImageId

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(imageId)

	return resourceZenlayerCloudImageRead(ctx, d, meta)
}

func resourceZenlayerCloudImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "resource.zenlayer_image.read")()

	var diags diag.Diagnostics

	imageId := d.Id()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var imageInfo *vm.DescribeImageResponseParams
	var errRet error

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		imageInfo, errRet = vmService.DescribeImageById(ctx, imageId)
		if errRet != nil {
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if imageInfo == nil {
			return nil
		}
		if imageInfo.ImageStatus == VmImageStatusUnavailable {
			return resource.NonRetryableError(fmt.Errorf("status of image (%s) is not available", imageId))
		}
		if imageInfo.ImageStatus == VmImageStatusCreating {
			return resource.RetryableError(fmt.Errorf("waiting for image %s operation", imageId))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if imageInfo == nil {
		d.SetId("")
		tflog.Info(ctx, "image not exist", map[string]interface{}{
			"imageId": imageId,
		})
		return nil
	}

	// image info
	_ = d.Set("image_name", imageInfo.ImageName)
	_ = d.Set("image_description", imageInfo.ImageDescription)
	imageSize, _ := strconv.Atoi(imageInfo.ImageSize)
	_ = d.Set("image_size", common.Integer(imageSize))

	return diags

}

func BuildImageState(vmService VmService, imageId string, ctx context.Context, d *schema.ResourceData) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{
			BmcEipStatusCreating,
		},
		Target: []string{
			ImageStatusAvailable,
			VmImageStatusUnavailable,
		},
		Refresh:        vmService.ImageStateRefreshFunc(ctx, imageId),
		Timeout:        d.Timeout(schema.TimeoutRead) - time.Minute,
		Delay:          10 * time.Second,
		MinTimeout:     5 * time.Second,
		NotFoundChecks: 3,
	}
}
