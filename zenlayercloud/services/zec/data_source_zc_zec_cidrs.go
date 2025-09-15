package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudCidrs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudCidrsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "IDs of the public CIDR block to be queried.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region ID that the public CIDR block locates at.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the public CIDR block list returned.",
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
			"cidrs": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of CIDR blocks. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the public CIDR block.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region ID that the public CIDR block locates at.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the public CIDR block.",
						},
						"cidr_block_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "CIDR block address.",
						},
						"netmask": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The IDR block size.",
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
							Description: "The type of CIDR block. Valid values: `Console`(for normal public CIDR), `BYOIP`(for bring your own IP).",
						},
						"used_ip_num": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Quantity of used CIDR IPs.",
						},
						"network_type": {
							Type:          schema.TypeString,
							Computed:      true,
							Description:   "Network types of public CIDR block. Valid values: `CN2Line`, `LocalLine`, `ChinaTelecom`, `ChinaUnicom`, `ChinaMobile`, `Cogent`.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the elastic IP.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the elastic IP.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudCidrsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_cidrs.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &CidrFilter{}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			filter.Ids = common.ToStringList(ids)
		}
	}

	if v, ok := d.GetOk("region_id"); ok {
		filter.RegionId = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = v.(string)
	}


	var nameRegex *regexp.Regexp

	if v, ok := d.GetOk("name_regex"); ok {
		imageName := v.(string)
		if imageName != "" {
			reg, err := regexp.Compile(imageName)
			if err != nil {
				return diag.Errorf("image_name_regex format error,%s", err.Error())
			}
			nameRegex = reg
		}
	}

	var cidrs []*zec.CidrInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		cidrs, e = zecService.DescribeCidrsByFilter(ctx, filter)
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
		if nameRegex != nil && !nameRegex.MatchString(*cidr.Name) {
			continue
		}
		mapping := map[string]interface{}{
			"id":                  cidr.CidrId,
			"region_id":           cidr.RegionId,
			"name":                cidr.Name,
			"cidr_block_address":  cidr.CidrBlock,
			"netmask":             cidr.Netmask,
			"resource_group_id":   cidr.ResourceGroupId,
			"resource_group_name": cidr.ResourceGroupName,
			"type":                cidr.Source,
			"used_ip_num":         cidr.UsedCount,
			"network_type":        cidr.EipV4Type,
			"create_time":         cidr.CreateTime,
			"status":              cidr.Status,
		}

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

type CidrFilter struct {
	Ids              []string
	RegionId         string
	ResourceGroupId  string
}
