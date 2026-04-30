/*
Use this resource to create a custom image from a ZEC instance, or import an
existing custom image to manage its name / tags via Terraform.

Example Usage

```hcl
resource "zenlayercloud_zec_image" "foo" {
  instance_id = "1660545330971680835"
  image_name  = "my-custom-image"

  tags = {
    env = "test"
  }
}
```

Import

A custom image can be imported using its id, e.g.

```
$ terraform import zenlayercloud_zec_image.foo <imageId>
```
*/
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

func ResourceZenlayerCloudZecImage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecImageCreate,
		ReadContext:   resourceZenlayerCloudZecImageRead,
		UpdateContext: resourceZenlayerCloudZecImageUpdate,
		DeleteContext: resourceZenlayerCloudZecImageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "ID of the ZEC instance to create the image from. Required for creation; ignored on import.",
			},
			"image_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(2, 63),
				Description:  "Name of the image. 2-63 chars; letters, digits, `-`, `_`, `.`; must start and end with a letter or digit.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Resource group the image belongs to. Defaults to the default resource group.",
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Tags bound to the image.",
			},

			"image_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the image. Typically `CUSTOM_IMAGE` for resources created here.",
			},
			"image_source": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Source of the image.",
			},
			"image_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the image.",
			},
			"image_size": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Size of the image, in GiB.",
			},
			"image_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "OS version of the image.",
			},
			"image_description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the image.",
			},
			"category": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Image catalog, e.g. `CentOS`, `Ubuntu`.",
			},
			"os_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "OS type of the image, such as `windows` or `linux`.",
			},
			"nic_network_type": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Supported NIC network types.",
			},
		},
	}
}

func resourceZenlayerCloudZecImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	instanceId, ok := d.GetOk("instance_id")
	if !ok || instanceId.(string) == "" {
		return diag.Errorf("`instance_id` is required to create a custom image")
	}

	request := zec.NewCreateImageRequest()
	request.InstanceId = common.String(instanceId.(string))
	request.ImageName = common.String(d.Get("image_name").(string))
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

	var imageId string
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		response, err := zecService.client.WithZec2Client().CreateImage(request)
		defer common2.LogApiRequest(ctx, "CreateImage", request, response, err)
		if err != nil {
			return common2.RetryError(ctx, err, common2.InternalServerError, common.NetworkError)
		}
		if response == nil || response.Response == nil || response.Response.ImageId == nil {
			return resource.NonRetryableError(fmt.Errorf("CreateImage returned empty imageId"))
		}
		imageId = *response.Response.ImageId
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(imageId)

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ZecImageStatusCreating,
			ZecImageStatusProcessing,
			ZecImageStatusSyncing,
		},
		Target: []string{
			ZecImageStatusAvailable,
		},
		Refresh:    zecService.ImageStateRefreshFunc(ctx, imageId, []string{ZecImageStatusFailed, ZecImageStatusUnavailable}),
		Timeout:    d.Timeout(schema.TimeoutCreate) - time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for image (%s) to become AVAILABLE: %v", imageId, err))
	}

	return resourceZenlayerCloudZecImageRead(ctx, d, meta)
}

func resourceZenlayerCloudZecImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	imageId := d.Id()

	var image *zec.CustomImage
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		image, e = zecService.DescribeImageById(ctx, imageId)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError, common.NetworkError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if image == nil {
		tflog.Info(ctx, "image not found, removing from state", map[string]interface{}{"imageId": imageId})
		d.SetId("")
		return nil
	}

	_ = d.Set("image_name", image.ImageName)
	_ = d.Set("resource_group_id", image.ResourceGroupId)
	_ = d.Set("image_type", image.ImageType)
	_ = d.Set("image_source", image.ImageSource)
	_ = d.Set("image_status", image.ImageStatus)
	_ = d.Set("image_size", image.ImageSize)
	_ = d.Set("image_version", image.ImageVersion)
	_ = d.Set("image_description", image.ImageDescription)
	_ = d.Set("category", image.Category)
	_ = d.Set("os_type", image.OsType)
	_ = d.Set("nic_network_type", image.NicNetworkType)

	if image.Tags != nil {
		tagMap, errRet := common2.TagsToMap(image.Tags)
		if errRet != nil {
			return diag.FromErr(errRet)
		}
		_ = d.Set("tags", tagMap)
	}

	return nil
}

func resourceZenlayerCloudZecImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	imageId := d.Id()
	d.Partial(true)

	if d.HasChange("image_name") {
		err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate)-time.Minute, func() *resource.RetryError {
			request := zec.NewModifyImagesAttributesRequest()
			request.ImageIds = []string{imageId}
			request.ImageName = common.String(d.Get("image_name").(string))
			response, err := zecService.client.WithZec2Client().ModifyImagesAttributes(request)
			defer common2.LogApiRequest(ctx, "ModifyImagesAttributes", request, response, err)
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
			request.Resources = []string{imageId}
			response, err := zecService.client.WithUsrClient().AddResourceResourceGroup(request)
			defer common2.LogApiRequest(ctx, "AddResourceResourceGroup", request, response, err)
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
		if err := zrmService.ModifyResourceTags(ctx, d, imageId); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)
	return resourceZenlayerCloudZecImageRead(ctx, d, meta)
}

func resourceZenlayerCloudZecImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	imageId := d.Id()

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		errRet := zecService.DeleteImage(ctx, imageId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if ok && ee.Code == common2.ResourceNotFound {
				return nil
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError, common.NetworkError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete)-time.Minute, func() *resource.RetryError {
		image, errRet := zecService.DescribeImageById(ctx, imageId)
		if errRet != nil {
			ee, ok := errRet.(*common.ZenlayerCloudSdkError)
			if ok && ee.Code == common2.ResourceNotFound {
				return nil
			}
			return common2.RetryError(ctx, errRet, common2.InternalServerError)
		}
		if image == nil {
			return nil
		}
		if image.ImageStatus != nil && *image.ImageStatus == ZecImageStatusDeleting {
			return resource.RetryableError(fmt.Errorf("waiting for image %s to be deleted, current status: %s", imageId, *image.ImageStatus))
		}
		return resource.RetryableError(fmt.Errorf("image %s still exists, current status: %v", imageId, image.ImageStatus))
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
