/*
Use this data source to get all vm available zones.

Example Usage

```hcl
data "zenlayercloud_zvm_zones" "all" {
}

data "zenlayercloud_zvm_zones" "sel" {
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
	vm "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/vm20230313"
	"regexp"
)

func dataSourceZenlayerCloudZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZonesRead,

		Schema: map[string]*schema.Schema{
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
				Description:  "A regex string to apply to the zone list returned.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"zones": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of availability zone. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the zone, such as `FRA-A`.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the zone, like `Frankfurt`, usually not used in api parameter.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zvm_zones.read")()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("image_name_regex format error,%s", errRet.Error())
		}
	}

	var zones []*vm.ZoneInfo
	err := resource.RetryContext(ctx, common.ReadRetryTimeout, func() *resource.RetryError {
		zones, errRet = vmService.DescribeZones(ctx)
		if errRet != nil {
			return common.RetryError(ctx, errRet, common.InternalServerError, common.ReadTimedOut)
		}
		return nil
	})
	zoneList := make([]map[string]interface{}, 0, len(zones))
	ids := make([]string, 0, len(zones))
	for _, zone := range zones {
		if nameRegex != nil && !nameRegex.MatchString(zone.ZoneId) {
			continue
		}

		mapping := map[string]interface{}{
			"name":        zone.ZoneId,
			"description": zone.ZoneName,
		}
		zoneList = append(zoneList, mapping)
		ids = append(ids, zone.ZoneId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("zones", zoneList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), zoneList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
