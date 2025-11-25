package zdns

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	common2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	pvtdns "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zdns20251101"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudPvtdnsZones() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudPvtdnsZonesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "IDs of the private DNS zones to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the private DNS zone list returned.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Resource group ID.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"zones": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of private DNS zones. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the private DNS zone.",
						},
						"zone_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the private DNS zone.",
						},
						"proxy_pattern": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Indicate whether the recursive resolution proxy is enabled or disabled.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of resource group.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resource group name.",
						},
						"tags": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "tags.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tag_key": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "tag key.",
									},
									"tag_value": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "tag value.",
									},
								},
							},
						},
						"remark": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Remark of the private DNS zone.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the private DNS zone.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudPvtdnsZonesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_pvtdns_zones.read")()

	pvtDnsService := ZdnsService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &PrivateZoneFilter{}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			filter.Ids = common.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = v.(string)
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		name := v.(string)
		if name != "" {
			reg, err := regexp.Compile(name)
			if err != nil {
				return diag.Errorf("name_regex format error: %s", err.Error())
			}
			nameRegex = reg
		}
	}

	var zones []*pvtdns.PrivateZone
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		zones, e = pvtDnsService.DescribePrivateZonesByFilter(ctx, filter)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0)
	zoneList := make([]map[string]interface{}, 0)

	for _, zone := range zones {
		if nameRegex != nil && !nameRegex.MatchString(*zone.ZoneName) {
			continue
		}
		mapping := map[string]interface{}{
			"id":              zone.ZoneId,
			"zone_name":       zone.ZoneName,
			"create_time":     zone.CreateTime,
			"resource_group_id":  zone.ResourceGroup.ResourceGroupId,
			"resource_group_name": zone.ResourceGroup.ResourceGroupName,
			"remark":          zone.Remark,
			"proxy_pattern": zone.ProxyPattern,
		}

		// Handle tags
		if zone.Tags != nil {
			tags := make(map[string]string)
			for _, tag := range zone.Tags.Tags {
				tags[*tag.Key] = common2.ToString(tag.Value)
			}
			mapping["tags"] = tags
		}

		zoneList = append(zoneList, mapping)
		ids = append(ids, *zone.ZoneId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	_ = d.Set("zones", zoneList)

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), zoneList); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

type PrivateZoneFilter struct {
	Ids              []string
	ResourceGroupId  string
}
