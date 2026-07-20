package zec

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
)

func DataSourceZenlayerCloudZecGpuInstanceConfigs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecGpuInstanceConfigsRead,

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Zone ID to filter GPU instance configs, e.g. `na-us-la-2a`. Returns all zones if not specified.",
			},
			"instance_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "GPU instance type to filter, e.g. `z3a.g.C49.c8m32.1`. Returns all types if not specified.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"data_set": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of GPU instance type quota items. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Zone ID, e.g. `na-us-la-2a`.",
						},
						"instance_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "GPU instance type ID, e.g. `z3a.g.C49.c8m32.1`. Pass this value to `instance_type` when creating or modifying an instance.",
						},
						"instance_type_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Human-readable description of the GPU instance type.",
						},
						"cpu_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of vCPUs.",
						},
						"memory": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Memory size in GiB.",
						},
						"gpu_amount": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of GPU cards.",
						},
						"bps": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum bandwidth per NIC in bits per second.",
						},
						"pps": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum packets per second per NIC.",
						},
						"inventory_capacity": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Inventory capacity level for this GPU instance type.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecGpuInstanceConfigsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_gpu_instance_configs.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var zoneId, instanceType string
	if v, ok := d.GetOk("zone_id"); ok {
		zoneId = v.(string)
	}
	if v, ok := d.GetOk("instance_type"); ok {
		instanceType = v.(string)
	}

	var items []map[string]interface{}
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		results, e := zecService.DescribeGpuInstanceConfigInfos(ctx, zoneId, instanceType)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		ids := make([]string, 0, len(results))
		items = make([]map[string]interface{}, 0, len(results))
		for _, item := range results {
			entry := map[string]interface{}{}
			if item.ZoneId != nil {
				entry["zone_id"] = *item.ZoneId
				ids = append(ids, *item.ZoneId)
			}
			if item.InstanceType != nil {
				entry["instance_type"] = *item.InstanceType
			}
			if item.InstanceTypeName != nil {
				entry["instance_type_name"] = *item.InstanceTypeName
			}
			if item.CpuCount != nil {
				entry["cpu_count"] = *item.CpuCount
			}
			if item.Memory != nil {
				entry["memory"] = *item.Memory
			}
			if item.GpuAmount != nil {
				entry["gpu_amount"] = *item.GpuAmount
			}
			if item.Bps != nil {
				entry["bps"] = int(*item.Bps)
			}
			if item.Pps != nil {
				entry["pps"] = int(*item.Pps)
			}
			if item.InventoryCapacity != nil {
				entry["inventory_capacity"] = *item.InventoryCapacity
			}
			items = append(items, entry)
		}
		d.SetId(common2.DataResourceIdHash(ids))
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("data_set", items); err != nil {
		return diag.FromErr(err)
	}

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), items); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
