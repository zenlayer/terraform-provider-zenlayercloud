/*
Use this data source to get the available regions for vpc.

Example Usage

```hcl
data "zenlayercloud_bmc_vpc_regions" "sel-region" {
  availability_zone = "SEL-A"
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"time"
)

func dataSourceZenlayerCloudVpcRegions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudVpcRegionsRead,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The zone that the vpc region contains.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region that the vpc locates at.",
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
				Description: "An information list of vpc regions. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the region.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the region.",
						},
						"availability_zones": {
							Type:        schema.TypeSet,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The zones that the vpc region contains.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudVpcRegionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_bmc_vpc_regions.read")()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	request := bmc.NewDescribeVpcAvailableRegionsRequest()
	if v, ok := d.GetOk("region"); ok {
		request.VpcRegionId = v.(string)

	}
	if v, ok := d.GetOk("availability_zone"); ok {
		request.ZoneId = v.(string)
	}

	var regions []*bmc.VpcRegion
	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		var response *bmc.DescribeVpcAvailableRegionsResponse
		response, e = bmcService.client.WithBmcClient().DescribeVpcAvailableRegions(request)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError, common.ReadTimedOut)
		}

		regions = response.Response.VpcRegionSet
		return nil
	}); err != nil {
		diag.FromErr(err)
	}

	regionList := make([]map[string]interface{}, 0, len(regions))
	ids := make([]string, 0, len(regions))
	for _, region := range regions {
		mapping := map[string]interface{}{
			"availability_zones": region.ZoneIds,
			"id":                 region.VpcRegionId,
			"name":               region.VpcRegionName,
		}
		regionList = append(regionList, mapping)
		ids = append(ids, region.VpcRegionId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err := d.Set("regions", regionList)
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
