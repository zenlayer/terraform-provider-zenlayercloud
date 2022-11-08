/*
Use this data source to query images.

Example Usage

```hcl
data "zenlayercloud_bmc_images" "foo" {
	catalog = "centos"
    instance_type_id = "S9I"
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"time"
)

func dataSourceZenlayerCloudImages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudImagesRead,

		Schema: map[string]*schema.Schema{
			"image_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the image.",
			},
			"image_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ImageTypes, false),
				Description:  "The image type. Valid values: 'PUBLIC_IMAGE', 'CUSTOM_IMAGE'.",
			},
			"image_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the image, such as `CentOS7.4-x86_64`.",
			},
			"catalog": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ImageCatalogs, false),
				Description:  "The catalog which the image belongs to. Valid values: 'centos', 'windows', 'ubuntu', 'debian', 'esxi'.",
			},
			"os_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(OsTypes, false),
				Description:  "os type of the image. Valid values: 'windows', 'linux'.",
			},
			"instance_type_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter images which are supported to install on specified instance type, such as `M6C`.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"images": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of image. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the image.",
						},
						"image_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the image. with value: `PUBLIC_IMAGE` and `CUSTOM_IMAGE`.",
						},
						"catalog": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Created time of the image.",
						},
						"image_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the image.",
						},
						"os_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the image, windows or linux.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudImagesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "data_source.zenlayercloud_bmc_images.read")()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := bmc.NewDescribeImagesRequest()

	if v, ok := d.GetOk("image_id"); ok {
		if v != "" {
			request.ImageIds = []string{v.(string)}
		}
	}
	if v, ok := d.GetOk("image_name"); ok {
		request.ImageName = v.(string)
	}

	if v, ok := d.GetOk("os_type"); ok {
		request.OsType = v.(string)
	}

	if v, ok := d.GetOk("image_type"); ok {
		request.ImageType = v.(string)
	}

	if v, ok := d.GetOk("instance_type_id"); ok {
		request.InstanceTypeId = v.(string)
	}

	if v, ok := d.GetOk("catalog"); ok {
		request.Catalog = v.(string)
	}

	var images []*bmc.ImageInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		var response *bmc.DescribeImagesResponse
		response, e = bmcService.client.WithBmcClient().DescribeImages(request)
		logApiRequest(ctx, "DescribeImages", request, response, e)
		if e != nil {
			return retryError(ctx, e, InternalServerError)
		}
		images = response.Response.Images
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	imageList := make([]map[string]interface{}, 0, len(images))
	ids := make([]string, 0, len(images))
	for _, image := range images {
		mapping := map[string]interface{}{
			"image_id":   image.ImageId,
			"os_type":    image.OsType,
			"image_type": image.ImageType,
			"catalog":    image.Catalog,
			"image_name": image.ImageName,
		}
		imageList = append(imageList, mapping)
		ids = append(ids, image.ImageId)
	}

	d.SetId(dataResourceIdHash(ids))
	err = d.Set("images", imageList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), imageList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
