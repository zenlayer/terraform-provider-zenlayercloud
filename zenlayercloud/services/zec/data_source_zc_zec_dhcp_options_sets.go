package zec

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
)

type DhcpOptionsSetFilter struct {
	dhcpOptionsSetIds []string
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
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IPv4 lease time.",
						},
						"ipv6_lease_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IPv6 lease time.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the DHCP options set.",
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
		if nameRegex != nil && !nameRegex.MatchString(*dhcpOptionsSet.DhcpOptionsName) {
			continue
		}

		ids = append(ids, *dhcpOptionsSet.DhcpOptionsSetId)
		dhcpOptionsSetMap := map[string]interface{}{
			"id":                       dhcpOptionsSet.DhcpOptionsSetId,
			"domain_name_servers":      dhcpOptionsSet.DomainNameServers,
			"ipv6_domain_name_servers": dhcpOptionsSet.Ipv6DomainNameServers,
			"lease_time":               dhcpOptionsSet.LeaseTime,
			"ipv6_lease_time":          dhcpOptionsSet.Ipv6LeaseTime,
			"description":              dhcpOptionsSet.Description,
			"create_time":              dhcpOptionsSet.CreateTime,
		}
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
