package zec

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func DataSourceZenlayerCloudZecImages() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecImagesRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Zone of the images to be queried.",
			},
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the image to be queried.",
			},
			"image_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ImageTypes, false),
				Description:  "The image type. Valid values: 'PUBLIC_IMAGE', 'CUSTOM_IMAGE'.",
			},
			"image_status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"CREATING", "AVAILABLE", "UNAVAILABLE", "SYNCING", "FAILED", "DELETING", "PROCESSING"}, false),
				Description:  "Filter by image status. Valid values: 'CREATING', 'AVAILABLE', 'UNAVAILABLE', 'SYNCING', 'FAILED', 'DELETING', 'PROCESSING'.",
			},
			"image_source": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{ZecImageSourceOfficial, ZecImageSourceMarketing, ZecImageSourceCustom}, false),
				Description:  "Filter by image source. Valid values: 'OFFICIAL', 'MARKETING', 'CUSTOM'.",
			},
			"image_name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
				Description:  "A regex string to apply to the image list returned by ZenlayerCloud, conflict with 'os_name'. **NOTE**: it is not wildcard, should look like `image_name_regex = \"^CentOS\\s+6\\.8\\s+64\\w*\"`.",
			},
			"category": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The catalog which the image belongs to. such as `CentOS`, `Windows`, `FreeBSD` etc.",
			},
			"os_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"windows", "linux", "bsd", "android", "any"}, false),
				Description:  "os type of the image. Valid values: 'windows', 'linux', 'bsd', 'android', 'any'.",
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
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the image.",
						},
						"image_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the image. With value: `PUBLIC_IMAGE` and `CUSTOM_IMAGE`.",
						},
						"category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The catalog which the image belongs to. With values: 'CentOS', 'Windows', 'Ubuntu', 'Debian'.",
						},
						"image_version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The version of image, such as 'Server 20.04 LTS'.",
						},
						"image_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the image.",
						},
						"os_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the image, `windows` or `linux`.",
						},
						"image_size": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The size of image. Measured in GiB.",
						},
						"image_description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of image.",
						},
						"image_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of image. Values: 'CREATING', 'AVAILABLE', 'UNAVAILABLE', 'SYNCING', 'FAILED', 'DELETING'.",
						},
						"image_source": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The source of image.",
						},
						"nic_network_type": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The supported NIC network types of image. Each element is one of: `auto` (adaptive), `vf,virtio` (failover), `virtio` (software emulation), `vf` (physical passthrough).",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The available tags within this image.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecImagesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_images.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &ImageFilter{
		zoneId: d.Get("availability_zone").(string),
	}

	if v, ok := d.GetOk("ids"); ok {
		vpcIds := v.(*schema.Set).List()
		if len(vpcIds) > 0 {
			filter.imageIds = common.ToStringList(vpcIds)
		}
	}

	var imageNameRegex *regexp.Regexp

	if v, ok := d.GetOk("image_name_regex"); ok {
		imageName := v.(string)
		if imageName != "" {
			reg, err := regexp.Compile(imageName)
			if err != nil {
				return diag.Errorf("image_name_regex format error,%s", err.Error())
			}
			imageNameRegex = reg
		}
	}
	if v, ok := d.GetOk("os_type"); ok {
		filter.osType = v.(string)
	}

	if v, ok := d.GetOk("image_type"); ok {
		filter.imageType = v.(string)
	}

	if v, ok := d.GetOk("category"); ok {
		filter.category = v.(string)
	}

	if v, ok := d.GetOk("image_status"); ok {
		filter.imageStatus = v.(string)
	}

	if v, ok := d.GetOk("image_source"); ok {
		filter.imageSource = v.(string)
	}

	var images []*zec.Image

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		images, e = zecService.DescribeImagesByFilter(filter)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError, common.ReadTimedOut)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var results []*zec.Image

	if imageNameRegex != nil {
		for _, image := range images {
			if imageNameRegex.MatchString(*image.ImageName) {
				results = append(results, image)
				continue
			}
		}
	} else {
		results = images
	}

	imageList := make([]map[string]interface{}, 0, len(results))
	ids := make([]string, 0, len(images))
	for _, image := range results {
		mapping := map[string]interface{}{
			"id":                image.ImageId,
			"os_type":           image.OsType,
			"image_type":        image.ImageType,
			"category":          image.Category,
			"image_name":        image.ImageName,
			"image_version":     image.ImageVersion,
			"image_size":        image.ImageSize,
			"image_description": image.ImageDescription,
			"image_status":      image.ImageStatus,
			"image_source":      image.ImageSource,
			"nic_network_type":  image.NicNetworkType,
		}

		tagMap, errRet := common.TagsToMap(image.Tags)
		if errRet != nil {
			return diag.FromErr(errRet)
		}
		mapping["tags"] = tagMap
		imageList = append(imageList, mapping)
		ids = append(ids, *image.ImageId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("images", imageList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), imageList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type ImageFilter struct {
	zoneId      string
	imageIds    []string
	imageType   string
	category    string
	osType      string
	imageStatus string
	imageSource string
}
