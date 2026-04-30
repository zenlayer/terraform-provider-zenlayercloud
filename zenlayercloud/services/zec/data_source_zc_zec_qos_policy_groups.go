package zec

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	sdkcommon "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func DataSourceZenlayerCloudZecQosPolicyGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecQosPolicyGroupsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs of the QoS policy groups to be queried.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region ID to filter QoS policy groups.",
			},
			"resource_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A member resource ID (EIP, IPv6 or UNMANAGED egress IP console UUID) to filter groups containing this resource.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to filter QoS policy groups by name.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Resource group ID to filter QoS policy groups.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"result": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of QoS policy groups. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the QoS policy group.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region ID of the QoS policy group.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the QoS policy group.",
						},
						"bandwidth_limit": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The shared bandwidth limit in Mbps.",
						},
						"rate_limit_mode": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The rate limit mode.",
						},
						"member_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of members in the group.",
						},
						"members": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The member list of the group.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The resource ID of the member.",
									},
									"ip_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The IP type of the member.",
									},
								},
							},
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resource group ID.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the resource group.",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The tags of the QoS policy group.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the QoS policy group.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecQosPolicyGroupsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_qos_policy_groups.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &QosPolicyGroupFilter{}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			filter.Ids = common.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("region_id"); ok {
		filter.RegionId = v.(string)
	}

	if v, ok := d.GetOk("resource_id"); ok {
		filter.ResourceId = v.(string)
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

	var groups []*zec.QosPolicyGroup
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		groups, e = zecService.DescribeQosPolicyGroupsByFilter(ctx, filter)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0)
	groupList := make([]map[string]interface{}, 0)

	for _, group := range groups {
		if nameRegex != nil && !nameRegex.MatchString(sdkcommon.ToString(group.Name)) {
			continue
		}

		members := make([]map[string]interface{}, 0)
		for _, m := range group.Members {
			members = append(members, map[string]interface{}{
				"resource_id": sdkcommon.ToString(m.ResourceId),
				"ip_type":     sdkcommon.ToString(m.IpType),
			})
		}

		mapping := map[string]interface{}{
			"id":              group.QosPolicyGroupId,
			"region_id":       group.RegionId,
			"name":            group.Name,
			"bandwidth_limit": int(sdkcommon.ToInt64(group.BandwidthLimit)),
			"rate_limit_mode": group.RateLimitMode,
			"member_count":    group.MemberCount,
			"members":         members,
			"create_time":     group.CreateTime,
		}

		if group.ResourceGroup != nil {
			mapping["resource_group_id"] = group.ResourceGroup.ResourceGroupId
			mapping["resource_group_name"] = group.ResourceGroup.ResourceGroupName
		}

		tagMap, errRet := common.TagsToMap(group.Tags)
		if errRet != nil {
			return diag.FromErr(errRet)
		}
		mapping["tags"] = tagMap

		groupList = append(groupList, mapping)
		ids = append(ids, sdkcommon.ToString(group.QosPolicyGroupId))
	}

	d.SetId(common.DataResourceIdHash(ids))
	_ = d.Set("result", groupList)

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), groupList); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}
