package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudZecSubnets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecSubnetsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "ID of the subnets to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the subnet list returned.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region that the subnet locates at.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			"result": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of subnets. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the subnet.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the VPC to be associated.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the subnet.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region that the subnet locates at.",
						},
						"cidr_block": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IPv4 network segment.",
						},
						"ipv6_cidr_block": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IPv6 network segment.",
						},
						"ip_stack_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Subnet IP stack type. Values: `IPv4`, `IPv6`, `IPv4_IPv6`.",
						},
						"ipv6_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IPv6 type. Valid values: `Public`, `Private`.",
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether it is the default subnet.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the subnet.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecSubnetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_subnets.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	filter := &SubnetFilter{}

	if v, ok := d.GetOk("ids"); ok {
		subnetIds := v.(*schema.Set).List()
		if len(subnetIds) > 0 {
			filter.ids = common.ToStringList(subnetIds)
		}
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	if v, ok := d.GetOk("region_id"); ok {
		filter.RegionId = v.(string)
	}

	var subnets []*zec.SubnetInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		subnets, e = zecService.DescribeSubnetsByFilter(ctx, filter)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	subnetList := make([]map[string]interface{}, 0, len(subnets))
	ids := make([]string, 0, len(subnets))

	for _, subnet := range subnets {
		if nameRegex != nil && !nameRegex.MatchString(subnet.Name) {
			continue
		}

		mapping := map[string]interface{}{
			"id":       subnet.SubnetId,
			"vpc_id":          subnet.VpcId,
			"name":            subnet.Name,
			"region_id":       subnet.RegionId,
			"cidr_block":      subnet.CidrBlock,
			"ipv6_cidr_block": subnet.Ipv6CidrBlock,
			"ip_stack_type":   subnet.StackType,
			"ipv6_type":       subnet.Ipv6Type,
			"is_default":      subnet.IsDefault,
			"create_time":     subnet.CreateTime,
		}

		subnetList = append(subnetList, mapping)
		ids = append(ids, subnet.SubnetId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	_ = d.Set("result", subnetList)

	return nil
}

type SubnetFilter struct {
	ids      []string
	RegionId string
}
