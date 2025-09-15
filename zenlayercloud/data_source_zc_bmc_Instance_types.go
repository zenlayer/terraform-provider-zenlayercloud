/*
Use this data source to query instances types.

Example Usage

```hcl
data "zenlayercloud_bmc_instance_types" "foo" {

}

data "zenlayercloud_bmc_instance_types" "sel" {
  availability_zone    = "SEL-A"
  instance_charge_type = "PREPAID"
  exclude_sold_out     = true
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
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"time"
)

func dataSourceZenlayerCloudInstanceTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudInstanceTypesRead,

		Schema: map[string]*schema.Schema{
			"instance_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      BmcChargeTypePostpaid,
				ValidateFunc: validation.StringInSlice(BmcChargeTypes, false),
				Description:  "The charge type of instance. Valid values are `POSTPAID`, `PREPAID`. The default is `POSTPAID`.",
			},
			"instance_type_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The instance type id of the instance.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The available zone that the BMC instance locates at.",
			},
			"exclude_sold_out": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicate to filter instances types that is sold out or not, default is false.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed values.
			"instance_types": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of available bmc instance types. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The zone id that the bmc instance locates at.",
						},
						"instance_type_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type ID of the instance.",
						},
						"internet_charge_types": {
							Type:        schema.TypeSet,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The supported internet charge types of the instance at specified zone.",
						},
						"maximum_bandwidth_out": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum public bandwidth of the instance type.",
						},
						"default_traffic_package_size": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "The default value of traffic package size.",
						},
						"sell_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Sell status of the instance.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudInstanceTypesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_bmc_instance_types.read")()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := bmc.NewDescribeAvailableResourcesRequest()

	if v, ok := d.GetOk("instance_charge_type"); ok {
		if v != "" {
			request.InstanceChargeType = v.(string)
		}
	}
	if v, ok := d.GetOk("instance_type_id"); ok {
		request.InstanceTypeId = v.(string)
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		request.ZoneId = v.(string)
	}

	var excludeSoldOut bool
	if _, ok := d.GetOk("exclude_sold_out"); ok {
		excludeSoldOut = true
	}

	var instanceTypes []*bmc.AvailableResource

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		var response *bmc.DescribeAvailableResourcesResponse
		response, e = bmcService.client.WithBmcClient().DescribeAvailableResources(request)
		common.LogApiRequest(ctx, "DescribeAvailableResources", request, response, e)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		instanceTypes = response.Response.AvailableResources
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	instanceTypeList := make([]map[string]interface{}, 0, len(instanceTypes))
	ids := make([]string, 0, len(instanceTypes))
	for _, instanceType := range instanceTypes {
		if excludeSoldOut && instanceType.SellStatus == "SOLD_OUT" {
			continue
		}
		mapping := map[string]interface{}{
			"availability_zone":            instanceType.ZoneId,
			"instance_type_id":             instanceType.InstanceTypeId,
			"internet_charge_types":        instanceType.InternetChargeTypes,
			"maximum_bandwidth_out":        instanceType.MaximumBandwidthOut,
			"default_traffic_package_size": instanceType.DefaultTrafficPackageSize,
			"sell_status":                  instanceType.SellStatus,
		}

		instanceTypeList = append(instanceTypeList, mapping)
		ids = append(ids, instanceType.InstanceTypeId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("instance_types", instanceTypeList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), instanceTypeList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
