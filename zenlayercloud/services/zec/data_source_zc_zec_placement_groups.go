package zec

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func DataSourceZenlayerCloudZecPlacementGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecPlacementGroupsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the placement groups to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to filter results by placement group name.",
			},
			"zone_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Zone ID to filter placement groups.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group to filter placement groups.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"placement_group_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of placement groups. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the placement group.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the placement group.",
						},
						"zone_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Zone ID of the placement group.",
						},
						"partition_num": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of partitions.",
						},
						"affinity": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The affinity level.",
						},
						"instance_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of instances in the placement group.",
						},
						"instance_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "The list of instance IDs associated with the placement group.",
						},
						"constraint_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The constraint satisfaction status of the placement group.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group ID.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group name.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the placement group.",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The tags of the placement group.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecPlacementGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_placement_groups.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &PlacementGroupFilter{}

	if v, ok := d.GetOk("ids"); ok {
		pgIds := v.(*schema.Set).List()
		if len(pgIds) > 0 {
			filter.PlacementGroupIds = common2.ToStringList(pgIds)
		}
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		var errRet error
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	if v, ok := d.GetOk("zone_id"); ok {
		filter.ZoneId = common.String(v.(string))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = common.String(v.(string))
	}

	var placementGroups []*zec.PlacementGroupInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		placementGroups, e = zecService.DescribePlacementGroupsByFilter(ctx, filter)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	pgList := make([]map[string]interface{}, 0, len(placementGroups))
	ids := make([]string, 0, len(placementGroups))
	for _, pg := range placementGroups {
		if nameRegex != nil && !nameRegex.MatchString(*pg.Name) {
			continue
		}
		mapping := map[string]interface{}{
			"id":            pg.PlacementGroupId,
			"name":          pg.Name,
			"zone_id":       pg.ZoneId,
			"partition_num": pg.PartitionNum,
			"affinity":      pg.Affinity,
			"instance_count": pg.InstanceCount,
			"instance_ids":       pg.InstanceIds,
			"constraint_status":  pg.ConstraintStatus,
			"create_time":        pg.CreateTime,
		}

		if pg.ResourceGroup != nil {
			mapping["resource_group_id"] = pg.ResourceGroup.ResourceGroupId
			mapping["resource_group_name"] = pg.ResourceGroup.ResourceGroupName
		}

		tagMap, errRet := common2.TagsToMap(pg.Tags)
		if errRet != nil {
			return diag.FromErr(errRet)
		}
		mapping["tags"] = tagMap

		pgList = append(pgList, mapping)
		ids = append(ids, *pg.PlacementGroupId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("placement_group_list", pgList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), pgList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
