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
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
)

func DataSourceZenlayerCloudIpv6Cidrs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudIpv6CidrsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "IDs of the public IPv6 CIDR block to be queried.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region ID that the public IPv6 CIDR block locates at.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the public IPv6 CIDR block list returned.",
			},
			"cidr_block": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IPv6 CIDR block address to filter, e.g. `2400:8a00::/28`.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Resource group ID.",
			},
			"asn": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ASN number to filter the IPv6 CIDR block list. Only valid for `BYOIP` CIDR blocks.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"cidrs": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of IPv6 CIDR blocks. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the IPv6 CIDR block.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region ID that the IPv6 CIDR block locates at.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the IPv6 CIDR block.",
						},
						"cidr_block_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IPv6 CIDR block address.",
						},
						"netmask": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The IPv6 CIDR block size.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resource group ID.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Name of resource group.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of IPv6 CIDR block. Valid values: `Console`(for normal public CIDR), `BYOIP`(for bring your own IP).",
						},
						"network_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network types of the IPv6 CIDR block.",
						},
						"subnet_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Subnet IDs that the IPv6 CIDR block is associated with.",
						},
						"nic_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "vNIC IDs that the IPv6 CIDR block is associated with.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the IPv6 CIDR block.",
						},
						"expired_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expiration time of the IPv6 CIDR block.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the IPv6 CIDR block.",
						},
						"asn": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ASN number. Only meaningful when the IPv6 CIDR block source is `BYOIP`; returns `0` for non-BYOIP CIDR blocks (the underlying API returns null in that case, which Terraform renders as `0` due to the limitation that `TypeInt` cannot represent null).",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The available tags within this IPv6 CIDR block.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudIpv6CidrsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_ipv6_cidrs.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &Ipv6CidrFilter{}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			filter.Ids = common.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("region_id"); ok {
		filter.RegionId = v.(string)
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		filter.CidrBlock = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("asn"); ok {
		asn := v.(int)
		filter.Asn = &asn
	}

	var nameRegex *regexp.Regexp

	if v, ok := d.GetOk("name_regex"); ok {
		name := v.(string)
		if name != "" {
			reg, err := regexp.Compile(name)
			if err != nil {
				return diag.Errorf("name_regex format error,%s", err.Error())
			}
			nameRegex = reg
		}
	}

	var cidrs []*zec.Ipv6CidrInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		cidrs, e = zecService.DescribeIpv6CidrsByFilter(ctx, filter)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0)
	cidrList := make([]map[string]interface{}, 0)

	for _, cidr := range cidrs {
		if nameRegex != nil && cidr.Name != nil && !nameRegex.MatchString(*cidr.Name) {
			continue
		}
		mapping := map[string]interface{}{
			"id":                 cidr.CidrId,
			"region_id":          cidr.RegionId,
			"name":               cidr.Name,
			"cidr_block_address": cidr.CidrBlock,
			"netmask":            cidr.Netmask,
			"type":               cidr.Source,
			"network_type":       cidr.NetworkLineType,
			"subnet_ids":         cidr.SubnetIds,
			"nic_ids":            cidr.NicIds,
			"create_time":        cidr.CreateTime,
			"expired_time":       cidr.ExpiredTime,
			"status":             cidr.Status,
			"asn":                cidr.Asn,
		}

		if cidr.ResourceGroup != nil {
			mapping["resource_group_id"] = cidr.ResourceGroup.ResourceGroupId
			mapping["resource_group_name"] = cidr.ResourceGroup.ResourceGroupName
		}

		tagMap, errRet := common.TagsToMap(cidr.Tags)
		if errRet != nil {
			return diag.FromErr(errRet)
		}
		mapping["tags"] = tagMap

		cidrList = append(cidrList, mapping)
		ids = append(ids, *cidr.CidrId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	_ = d.Set("cidrs", cidrList)

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), cidrList); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

type Ipv6CidrFilter struct {
	Ids             []string
	RegionId        string
	CidrBlock       string
	ResourceGroupId string
	Asn             *int
}
