/*
Use this data source to get all sdn data centers available.

Example Usage

```hcl
data "zenlayercloud_sdn_datacenters" "all" {
}

data "zenlayercloud_sdn_datacenters" "sel" {
	name_regex = "SEL*"
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
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	sdn "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/sdn20230830"
	"regexp"
)

func dataSourceZenlayerCloudSdnDatacenters() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudSdnDatacentersRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
				Description:  "A regex string to apply to the datacenter list returned.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"datacenters": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of availability datacenter. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the datacenter, which is a uuid format.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the datacenter, like `AP-Singapore1`, usually not used in api parameter.",
						},
						"address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The location of the datacenter.",
						},
						"city_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of city where the datacenter located, like `Singapore`.",
						},
						"country_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of country, like `Singapore`.",
						},
						"area_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region name, like `Asia Pacific`.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudSdnDatacentersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_datacenters.read")()

	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	var datacenters []*sdn.DatacenterInfo
	err := resource.RetryContext(ctx, common.ReadRetryTimeout, func() *resource.RetryError {
		datacenters, errRet = sdnService.DescribeDatacenters(ctx)
		if errRet != nil {
			return common.RetryError(ctx, errRet, common.InternalServerError, common.ReadTimedOut)
		}
		return nil
	})
	datacenterList := make([]map[string]interface{}, 0, len(datacenters))
	ids := make([]string, 0, len(datacenters))
	for _, datacenter := range datacenters {
		if nameRegex != nil && !nameRegex.MatchString(datacenter.DcName) {
			continue
		}

		mapping := map[string]interface{}{
			"id":           datacenter.DcId,
			"name":         datacenter.DcName,
			"address":      datacenter.DcAddress,
			"city_name":    datacenter.CityName,
			"country_name": datacenter.CountryName,
			"area_name":    datacenter.AreaName,
		}
		datacenterList = append(datacenterList, mapping)
		ids = append(ids, datacenter.DcId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("datacenters", datacenterList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), datacenterList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
