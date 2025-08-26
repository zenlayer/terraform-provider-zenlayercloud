/*
Use this data source to query instances type.

Example Usage

```hcl
data "zenlayercloud_instance_types" "foo" {

}

data "zenlayercloud_instance_types" "sel1c1g" {
  availability_zone = "SEL-A"
  cpu_count   		= 1
  memory	        = 1
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
	"time"
)

func dataSourceZenlayerCloudVmInstanceTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudVmInstanceTypesRead,

		Schema: map[string]*schema.Schema{
			"instance_charge_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      VmChargeTypePostpaid,
				ValidateFunc: validation.StringInSlice(VmChargeTypes, false),
				Description:  "The charge type of instance. Valid values are `POSTPAID`, `PREPAID`. The default is `POSTPAID`.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The available zone that the instance locates at.",
			},
			"instance_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The instance type of the instance.",
			},
			"cpu_count": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of CPU cores of the instance.",
			},
			"memory": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Instance memory capacity, unit in GB.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed values.
			"instance_type_quotas": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of zone available vm instance types. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The zone id that the vm instance locates at.",
						},
						"instance_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the instance.",
						},
						"internet_charge_type": {
							Type:        schema.TypeSet,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Internet charge type of the instance, Valid values are `ByBandwidth`, `ByTrafficPackage`, `ByInstanceBandwidth95` and `ByClusterBandwidth95`.",
						},
						"maximum_bandwidth_out": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum public bandwidth of the instance type.",
						},
						"cpu_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of CPU cores of the instance.",
						},
						"memory": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Instance memory capacity, unit in GB.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudVmInstanceTypesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zone_instance_config_infos.read")()

	vmService := VmService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := vm.NewDescribeZoneInstanceConfigInfosRequest()

	if v, ok := d.GetOk("instance_charge_type"); ok {
		if v != "" {
			request.InstanceChargeType = v.(string)
		}
	}
	if v, ok := d.GetOk("instance_type"); ok {
		request.InstanceType = v.(string)
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		request.ZoneId = v.(string)
	}

	cpu, cpuOk := d.GetOk("cpu_count")
	memory, memoryOk := d.GetOk("memory")

	var instanceTypes []*vm.InstanceTypeQuotaItem

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		var response *vm.DescribeZoneInstanceConfigInfosResponse
		response, e = vmService.client.WithVmClient().DescribeZoneInstanceConfigInfos(request)
		common.LogApiRequest(ctx, "DescribeZoneInstanceConfigInfos", request, response, e)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		instanceTypes = response.Response.InstanceTypeQuotaSet
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	instanceTypeList := make([]map[string]interface{}, 0, len(instanceTypes))
	ids := make([]string, 0, len(instanceTypes))
	for _, instanceType := range instanceTypes {
		flag := true
		if cpuOk && cpu.(int) != instanceType.CpuCount {
			flag = false
		}
		if memoryOk && memory.(int) != instanceType.Memory {
			flag = false
		}
		if flag {
			mapping := map[string]interface{}{
				"availability_zone":     instanceType.ZoneId,
				"instance_type":         instanceType.InstanceType,
				"internet_charge_type":  instanceType.InternetChargeTypes,
				"maximum_bandwidth_out": instanceType.InternetMaxBandwidthOutLimit,
				"cpu_count":             instanceType.CpuCount,
				"memory":                instanceType.Memory,
			}
			instanceTypeList = append(instanceTypeList, mapping)
			ids = append(ids, instanceType.InstanceType)
		}
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("instance_type_quotas", instanceTypeList)
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
