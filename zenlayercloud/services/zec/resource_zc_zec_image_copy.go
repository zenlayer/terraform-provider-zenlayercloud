/*
Use this resource to manage the cross-region distribution of a custom ZEC image.
Declare the full set of regions (including the source region) where the image
should exist; Terraform will call CopyImage for additions and DeleteImageCopy
for removals.

Example Usage

```hcl
resource "zenlayercloud_zec_image_copy" "example" {
  image_id       = zenlayercloud_zec_image.img.id
  region_id_list = ["na-west-1", "asia-southeast-1"]
}
```

Import

An image-copy resource can be imported using the image id, e.g.

```
$ terraform import zenlayercloud_zec_image_copy.example <imageId>
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
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	sdkcommon "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

func ResourceZenlayerCloudZecImageCopy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZenlayerCloudZecImageCopyCreate,
		ReadContext:   resourceZenlayerCloudZecImageCopyRead,
		UpdateContext: resourceZenlayerCloudZecImageCopyUpdate,
		DeleteContext: resourceZenlayerCloudZecImageCopyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"image_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the custom image to distribute.",
			},
			"region_id_list": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Full set of region IDs where the image should exist (including the source region). At least one region must remain at all times.",
			},
		},
	}
}

func resourceZenlayerCloudZecImageCopyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}
	imageId := d.Get("image_id").(string)

	desired := expandStringSet(d.Get("region_id_list").(*schema.Set))

	current, err := fetchRegionIdList(ctx, &zecService, imageId)
	if err != nil {
		return diag.FromErr(err)
	}

	toAdd := subtract(desired, current)
	if len(toAdd) > 0 {
		preConf := &resource.StateChangeConf{
			Pending:    []string{"PENDING"},
			Target:     []string{ZecImageStatusAvailable},
			Refresh:    zecService.ImageAvailableForCopyRefreshFunc(ctx, imageId),
			Timeout:    d.Timeout(schema.TimeoutCreate) - time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 5 * time.Second,
		}
		if _, err := preConf.WaitForStateContext(ctx); err != nil {
			return diag.FromErr(fmt.Errorf("image (%s) is not ready for copy: %v", imageId, err))
		}

		if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			e := zecService.CopyImage(ctx, imageId, toAdd)
			if e != nil {
				return common2.RetryError(ctx, e, common2.InternalServerError, sdkcommon.NetworkError)
			}
			return nil
		}); err != nil {
			return diag.FromErr(err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"PENDING"},
			Target:     []string{"COMPLETE"},
			Refresh:    zecService.ImageCopyStateRefreshFunc(ctx, imageId, desired, nil),
			Timeout:    d.Timeout(schema.TimeoutCreate) - time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 5 * time.Second,
		}
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for image (%s) copies to become available: %v", imageId, err))
		}
	}

	d.SetId(imageId)
	return resourceZenlayerCloudZecImageCopyRead(ctx, d, meta)
}

func resourceZenlayerCloudZecImageCopyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}
	imageId := d.Id()

	var regionIdList []string
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		regionIdList, e = fetchRegionIdList(ctx, &zecService, imageId)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError, sdkcommon.NetworkError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if regionIdList == nil {
		tflog.Info(ctx, "image not found, removing from state", map[string]interface{}{"imageId": imageId})
		d.SetId("")
		return nil
	}

	_ = d.Set("image_id", imageId)
	_ = d.Set("region_id_list", regionIdList)
	return nil
}

func resourceZenlayerCloudZecImageCopyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !d.HasChange("region_id_list") {
		return nil
	}

	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}
	imageId := d.Id()

	desired := expandStringSet(d.Get("region_id_list").(*schema.Set))

	current, err := fetchRegionIdList(ctx, &zecService, imageId)
	if err != nil {
		return diag.FromErr(err)
	}

	toAdd := subtract(desired, current)
	toRemove := subtract(current, desired)

	// Additions first: CopyImage causes SYNCING; DeleteImageCopy must wait until done.
	if len(toAdd) > 0 {
		preConf := &resource.StateChangeConf{
			Pending:    []string{"PENDING"},
			Target:     []string{ZecImageStatusAvailable},
			Refresh:    zecService.ImageAvailableForCopyRefreshFunc(ctx, imageId),
			Timeout:    d.Timeout(schema.TimeoutUpdate) - time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 5 * time.Second,
		}
		if _, err := preConf.WaitForStateContext(ctx); err != nil {
			return diag.FromErr(fmt.Errorf("image (%s) is not ready for copy: %v", imageId, err))
		}

		if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			e := zecService.CopyImage(ctx, imageId, toAdd)
			if e != nil {
				return common2.RetryError(ctx, e, common2.InternalServerError, sdkcommon.NetworkError)
			}
			return nil
		}); err != nil {
			return diag.FromErr(err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"PENDING"},
			Target:     []string{"COMPLETE"},
			Refresh:    zecService.ImageCopyStateRefreshFunc(ctx, imageId, toAdd, nil),
			Timeout:    d.Timeout(schema.TimeoutUpdate) - time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 5 * time.Second,
		}
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for image (%s) copies to become available: %v", imageId, err))
		}
	}

	if len(toRemove) > 0 {
		waitAvail := &resource.StateChangeConf{
			Pending:    []string{"PENDING"},
			Target:     []string{"COMPLETE"},
			Refresh: func() (interface{}, string, error) {
				image, e := zecService.DescribeCustomImageById(ctx, imageId)
				if e != nil {
					return nil, "", e
				}
				if image == nil {
					return nil, "NOT_FOUND", nil
				}
				if image.ImageStatus != nil && *image.ImageStatus == ZecImageStatusFailed {
					return image, ZecImageStatusFailed, fmt.Errorf("image %s in failed state", imageId)
				}
				if image.ImageStatus == nil || *image.ImageStatus != ZecImageStatusAvailable {
					return image, "PENDING", nil
				}
				return image, "COMPLETE", nil
			},
			Timeout:    d.Timeout(schema.TimeoutUpdate) - time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 5 * time.Second,
		}
		if _, err := waitAvail.WaitForStateContext(ctx); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for image (%s) to become AVAILABLE before delete: %v", imageId, err))
		}

		if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			e := zecService.DeleteImageCopy(ctx, imageId, toRemove)
			if e != nil {
				return common2.RetryError(ctx, e, common2.InternalServerError, sdkcommon.NetworkError)
			}
			return nil
		}); err != nil {
			return diag.FromErr(err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"PENDING"},
			Target:     []string{"COMPLETE"},
			Refresh:    zecService.ImageCopyStateRefreshFunc(ctx, imageId, nil, toRemove),
			Timeout:    d.Timeout(schema.TimeoutUpdate) - time.Minute,
			Delay:      5 * time.Second,
			MinTimeout: 5 * time.Second,
		}
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for image (%s) region removals to complete: %v", imageId, err))
		}
	}

	return resourceZenlayerCloudZecImageCopyRead(ctx, d, meta)
}

func resourceZenlayerCloudZecImageCopyDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

// fetchRegionIdList returns the current RegionIdList for the image, or nil if the image is not found.
func fetchRegionIdList(ctx context.Context, svc *ZecService, imageId string) ([]string, error) {
	image, err := svc.DescribeCustomImageById(ctx, imageId)
	if err != nil {
		return nil, err
	}
	if image == nil {
		return nil, nil
	}
	return image.RegionIdList, nil
}

// subtract returns elements in a that are not in b.
func subtract(a, b []string) []string {
	bSet := make(map[string]bool, len(b))
	for _, v := range b {
		bSet[v] = true
	}
	var result []string
	for _, v := range a {
		if !bSet[v] {
			result = append(result, v)
		}
	}
	return result
}

func expandStringSet(s *schema.Set) []string {
	list := s.List()
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}
