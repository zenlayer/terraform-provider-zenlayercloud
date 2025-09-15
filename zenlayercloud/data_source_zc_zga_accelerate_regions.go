/*
Use this data source to get all zga available accelerate regions by origin_region_id.

Example Usage
```hcl

	data "zenlayercloud_zga_origin_regions" "DE" {
		name_regex = "DE"
	}

	data "zenlayercloud_zga_accelerate_regions" "all" {
		origin_region_id = data.zenlayercloud_zga_origin_regions.DE.regions.0.id
	}

	data "zenlayercloud_zga_accelerate_regions" "FR" {
		origin_region_id = "FR"
		name_regex = "US*"
	}

```
*/
package zenlayercloud

import (
	"context"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"regexp"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zga "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zga20230706"
)

func dataSourceZenlayerCloudZgaAccelerateRegions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZgaAccelerateRegionsRead,
		Schema: map[string]*schema.Schema{
			"origin_region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the origin region, such as `FR`.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
				Description:  "A regex string to apply to the accelerate region list returned.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"regions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of availability accelerate region. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the region, such as `FR`.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the region, like `Frankfurt`, usually not used in api parameter.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZgaAccelerateRegionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zga_accelerate_regions.read")()

	var (
		nameRegex *regexp.Regexp
		errRet    error
	)
	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error ,%s", errRet.Error())
		}
	}

	var regions []*zga.Region
	err := resource.RetryContext(ctx, common.ReadRetryTimeout, func() *resource.RetryError {
		regions, errRet = NewZgaService(meta.(*connectivity.ZenlayerCloudClient)).
			DescribeAccelerateRegions(ctx, d.Get("origin_region_id").(string))
		if errRet != nil {
			return common.RetryError(ctx, errRet, common.InternalServerError, common.ReadTimedOut)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	regionList := make([]map[string]interface{}, 0, len(regions))
	ids := make([]string, 0, len(regions))
	for _, region := range regions {
		if nameRegex != nil && !nameRegex.MatchString(region.RegionId) {
			continue
		}
		regionList = append(regionList, map[string]interface{}{
			"id":          region.RegionId,
			"description": region.RegionName,
		})
		ids = append(ids, region.RegionId)
	}

	if len(regionList) == 0 {
		return diag.Errorf("Query returned no results. These regions may be closed or name_regex input wrong.")
	}

	sort.StringSlice(ids).Sort()

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("regions", regionList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), regionList); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
