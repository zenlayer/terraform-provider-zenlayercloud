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

func DataSourceZenlayerCloudZecHaVips() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecHaVipsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs of the HaVips to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to filter HaVips by name.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region ID where the HaVips are located.",
			},
			"vpc_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The VPC IDs to filter HaVips.",
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The subnet IDs to filter HaVips.",
			},
			"ip_addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The private IP addresses to filter HaVips.",
			},
			"instance_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Return HaVips that are bound to the specified instance IDs.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"result": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of HaVips. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the HaVip.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the HaVip.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region ID where the HaVip is located.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The VPC ID to which the HaVip belongs.",
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The subnet ID to which the HaVip belongs.",
						},
						"security_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The security group ID associated with the HaVip.",
						},
						"ip_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private IPv4 address of the HaVip.",
						},
						"associated_instances": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "The list of instance IDs associated with the HaVip.",
						},
						"master_instance_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The current master instance ID. Null when no instance is bound.",
						},
						"associated_eips": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The list of EIPs associated with the HaVip.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eip_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the EIP.",
									},
									"eip_address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The EIP address.",
									},
								},
							},
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The creation time of the HaVip.",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "The tags associated with the HaVip.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecHaVipsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_havips.read")()

	zecService := ZecService{client: meta.(*connectivity.ZenlayerCloudClient)}

	filter := &HaVipFilter{}

	if v, ok := d.GetOk("ids"); ok {
		filter.Ids = common.ToStringList(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("region_id"); ok {
		filter.RegionId = v.(string)
	}
	if v, ok := d.GetOk("vpc_ids"); ok {
		filter.VpcIds = common.ToStringList(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("subnet_ids"); ok {
		filter.SubnetIds = common.ToStringList(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("ip_addresses"); ok {
		filter.IpAddresses = common.ToStringList(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("instance_ids"); ok {
		filter.InstanceIds = common.ToStringList(v.(*schema.Set).List())
	}

	var nameRegex *regexp.Regexp
	if v, ok := d.GetOk("name_regex"); ok {
		if v.(string) != "" {
			reg, err := regexp.Compile(v.(string))
			if err != nil {
				return diag.Errorf("name_regex format error: %s", err.Error())
			}
			nameRegex = reg
		}
	}

	var haVips []*zec.HaVipInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		haVips, e = zecService.DescribeHaVipsByFilter(ctx, filter)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	ids := make([]string, 0)
	resultList := make([]map[string]interface{}, 0)

	for _, haVip := range haVips {
		if nameRegex != nil && !nameRegex.MatchString(sdkcommon.ToString(haVip.Name)) {
			continue
		}

		mapping := map[string]interface{}{
			"id":                   haVip.HaVipId,
			"name":                 haVip.Name,
			"region_id":            haVip.RegionId,
			"vpc_id":               haVip.VpcId,
			"subnet_id":            haVip.SubnetId,
			"security_group_id":    haVip.SecurityGroupId,
			"ip_address":           haVip.IpAddress,
			"associated_instances": haVip.AssociatedInstances,
			"master_instance_id":   haVip.MasterInstanceId,
			"create_time":          haVip.CreateTime,
		}

		associatedEips := make([]map[string]interface{}, 0, len(haVip.AssociatedEips))
		for _, eip := range haVip.AssociatedEips {
			associatedEips = append(associatedEips, map[string]interface{}{
				"eip_id":      eip.EipId,
				"eip_address": eip.EipAddress,
			})
		}
		mapping["associated_eips"] = associatedEips

		tagMap, err := common.TagsToMap(haVip.Tags)
		if err != nil {
			return diag.FromErr(err)
		}
		mapping["tags"] = tagMap

		resultList = append(resultList, mapping)
		ids = append(ids, *haVip.HaVipId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	_ = d.Set("result", resultList)

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), resultList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
