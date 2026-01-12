package zec

import (
	"context"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
)

type DhcpOptionsSetFilter struct {
	dhcpOptionsSetIds []string
	resourceGroupId   string
}

func DataSourceZenlayerCloudZecDhcpOptionsSets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecDhcpOptionsSetsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "List of DHCP options set IDs.",
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
				Description:  "Regular expression for DHCP options set names.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the resource group to filter DHCP options sets.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Output file path.",
			},
			"dhcp_options_sets": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of DHCP options sets.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "DHCP options set ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the DHCP options set.",
						},
						"domain_name_servers": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IPv4 DNS server IP.",
						},
						"ipv6_domain_name_servers": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IPv6 DNS server IP.",
						},
						"lease_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "IPv4 lease time.",
						},
						"ipv6_lease_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "IPv6 lease time.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the DHCP options set.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the resource group to which the DHCP options set belongs.",
						},
						"tags": {
							Type:        schema.TypeMap,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "Tags of the DHCP options set.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the DHCP options set.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecDhcpOptionsSetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	zenlayerCloudClient := meta.(*connectivity.ZenlayerCloudClient)
	zecService := &ZecService{client: zenlayerCloudClient}

	filter := &DhcpOptionsSetFilter{}
	if v, ok := d.GetOk("ids"); ok {
		list := v.(*schema.Set).List()
		if len(list) > 0 {
			filter.dhcpOptionsSetIds = common.ToStringList(list)
		}
	}

	// 处理资源组过滤
	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.resourceGroupId = v.(string)
	}

	var nameRegex *regexp.Regexp

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex = regexp.MustCompile(v.(string))
	}

	allDhcpOptionsSets, err := zecService.DescribeDhcpOptionsSetsByFilter(ctx, filter)
	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0, len(allDhcpOptionsSets))
	dhcpOptionsSets := make([]map[string]interface{}, 0, len(allDhcpOptionsSets))
	for _, dhcpOptionsSet := range allDhcpOptionsSets {
		if nameRegex != nil && !nameRegex.MatchString(*dhcpOptionsSet.DhcpOptionsSetName) {
			continue
		}

		ids = append(ids, *dhcpOptionsSet.DhcpOptionsSetId)

		dhcpOptionsSetMap := map[string]interface{}{
			"id":                       dhcpOptionsSet.DhcpOptionsSetId,
			"name":                     dhcpOptionsSet.DhcpOptionsSetName,
			"domain_name_servers":      dhcpOptionsSet.DomainNameServers,
			"ipv6_domain_name_servers": dhcpOptionsSet.Ipv6DomainNameServers,
			"lease_time":               0,
			"ipv6_lease_time":          0,
			"description":              dhcpOptionsSet.Description,
			"resource_group_id":        dhcpOptionsSet.ResourceGroupId,
			"create_time":              dhcpOptionsSet.CreateTime,
		}
		if dhcpOptionsSet.LeaseTime != nil {
			if leaseTime, err := strconv.Atoi(*dhcpOptionsSet.LeaseTime); err == nil {
				dhcpOptionsSetMap["lease_time"] = leaseTime
			}
		}
		if dhcpOptionsSet.Ipv6LeaseTime != nil {
			if ipv6LeaseTime, err := strconv.Atoi(*dhcpOptionsSet.Ipv6LeaseTime); err == nil {
				dhcpOptionsSetMap["ipv6_lease_time"] = ipv6LeaseTime
			}
		}
		// Read tags
		tags, err := common.TagsToMap(dhcpOptionsSet.Tags)
		if err != nil {
			return diag.FromErr(err)
		}
		dhcpOptionsSetMap["tags"] = tags

		dhcpOptionsSets = append(dhcpOptionsSets, dhcpOptionsSetMap)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("dhcp_options_sets", dhcpOptionsSets)
	if err != nil {
		return diag.FromErr(err)
	}

	if output, ok := d.GetOk("result_output_file"); ok && output != "" {
		if err := common.WriteToFile(output.(string), dhcpOptionsSets); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
