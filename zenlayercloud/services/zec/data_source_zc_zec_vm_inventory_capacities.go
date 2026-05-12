package zec

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func DataSourceZenlayerCloudZecVmInventoryCapacities() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecVmInventoryCapacitiesRead,

		Schema: map[string]*schema.Schema{
			"region_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "Node IDs to query, e.g. `asia-north-1`. Returns all nodes if not specified.",
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
				Description: "Inventory capacity list per node. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Node ID, e.g. `asia-north-1`.",
						},
						"capacity": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Overall inventory capacity level of the node. One of `LIMITED` (< 1000 cores), `NORMAL` (1000-2000 cores), `SUFFICIENT` (2000-5000 cores), `ABUNDANT` (>= 5000 cores).",
						},
						"instance_types": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Per-instance-type capacity breakdown. Entries with zero inventory are excluded.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"instance_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "CPU instance type, e.g. `z2a`, `z2i`, `z4a`.",
									},
									"gpu_spec": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "GPU model, e.g. `z4a.g.C49`. Only present for GPU instances.",
									},
									"capacity": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Inventory capacity level for this instance type. One of `LIMITED` (< 1000 cores), `NORMAL` (1000-2000 cores), `SUFFICIENT` (2000-5000 cores), `ABUNDANT` (>= 5000 cores).",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecVmInventoryCapacitiesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_vm_inventory_capacities.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	var regionIds []string
	if v, ok := d.GetOk("region_ids"); ok {
		regionIds = common2.ToStringList(v.(*schema.Set).List())
	}

	var dataset []*zec.VmRegionCapacityItem
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		dataset, e = zecService.DescribeVmInventoryCapacity(ctx, regionIds)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(dataset))
	dataSet := make([]map[string]interface{}, 0, len(dataset))
	for _, item := range dataset {
		instanceTypes := make([]map[string]interface{}, 0, len(item.InstanceTypes))
		for _, it := range item.InstanceTypes {
			gpuSpec := ""
			if it.GpuSpec != nil {
				gpuSpec = *it.GpuSpec
			}
			instanceTypes = append(instanceTypes, map[string]interface{}{
				"instance_type": it.InstanceType,
				"gpu_spec":      gpuSpec,
				"capacity":      it.Capacity,
			})
		}
		regionId := ""
		if item.RegionId != nil {
			regionId = *item.RegionId
			ids = append(ids, regionId)
		}
		dataSet = append(dataSet, map[string]interface{}{
			"region_id":      regionId,
			"capacity":       item.Capacity,
			"instance_types": instanceTypes,
		})
	}

	d.SetId(common2.DataResourceIdHash(ids))
	if err := d.Set("data_set", dataSet); err != nil {
		return diag.FromErr(err)
	}

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), dataSet); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
