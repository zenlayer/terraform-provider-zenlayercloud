package zlb

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zlb "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zlb20250401"
	"time"
)

func DataSourceZenlayerCloudZlbRegions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZlbRegionsRead,

		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of region that the load balancer locates at.",
			},
			"city_code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The code of the city where the region is located to be queried.",
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
				Description: "An information list of instances. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of region that support for load balancer. such as `asia-east-1`.",
						},
						"city_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the city where the region is located. such as `Shanghai`.",
						},
						"city_code": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The code of the city where the region is located. such as `SHA`.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZlbRegionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zlb_regions.read")()

	zlbService := ZlbService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var regionId = ""
	if v, ok := d.GetOk("region_id"); ok {
		regionId = v.(string)
	}

	var cityCode = ""
	if v, ok := d.GetOk("city_code"); ok {
		cityCode = v.(string)
	}

	var regions []*zlb.Region

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		regions, e = zlbService.DescribeLoadBalancerRegions()
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}
	regionList := make([]map[string]interface{}, 0, len(regions))
	ids := make([]string, 0, len(regions))
	for _, r := range regions {
		if regionId != "" && *r.RegionId != regionId {
			continue
		}
		if cityCode != "" && *r.CityCode != cityCode {
			continue
		}
		mapping := map[string]interface{}{
			"region_id": r.RegionId,
			"city_name": r.CityName,
			"city_code": r.CityCode,
		}

		regionList = append(regionList, mapping)
		ids = append(ids, *r.RegionId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("regions", regionList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), regionList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
