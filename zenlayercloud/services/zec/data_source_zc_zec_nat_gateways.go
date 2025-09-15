package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec2 "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"regexp"
	"time"
)

func DataSourceZenlayerCloudZecNatGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecNatGatewayRead,

		Schema: map[string]*schema.Schema{
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Region of the NAT gateway to be queried.",
			},
			"ids": {
				Type:        schema.TypeSet,
				Elem: &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "ids of the NAT gateway to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the NAT gateway list returned.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group grouped NAT gateway to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"nats": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of NAT gateways. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"nat_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the NAT gateway.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the NAT gateway.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region that the NAT gateway locates at.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the VPC to be associated.",
						},
						"subnet_ids": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "IDs of the subnets to be associated. if this value not set.",
						},
						"is_all_subnets": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether all the subnets of region is assigned to NAT gateway.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The status of NAT gateway.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group id that the NAT gateway belongs to.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group name that the NAT gateway belongs to.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the NAT gateway.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecNatGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_nat_gateways.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := &ZecNatGatewayFilter{}

	if v, ok := d.GetOk("region_id"); ok {
		request.RegionId = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		if len(ids) > 0 {
			request.Ids = common2.ToStringList(ids)
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

	var result []*zec2.NatGateway

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		result, e = zecService.DescribeNatGateways(ctx, request)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	nats := make([]map[string]interface{}, 0, len(result))

	ids := make([]string, 0, len(result))
	for _, natGateway := range result {
		if nameRegex != nil && !nameRegex.MatchString(*natGateway.Name) {
			continue
		}

		mapping := map[string]interface{}{
			"region_id":           natGateway.RegionId,
			"nat_id":              natGateway.NatGatewayId,
			"name":                natGateway.Name,
			"vpc_id":              natGateway.VpcId,
			"is_all_subnets":      natGateway.IsAllSubnets,
			"subnet_ids":          natGateway.SubnetIds,
			"create_time":         natGateway.CreateTime,
			"status":              natGateway.Status,
			"resource_group_id":   natGateway.ResourceGroupId,
			"resource_group_name": natGateway.ResourceGroupName,
		}
		nats = append(nats, mapping)
		ids = append(ids, *natGateway.NatGatewayId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("nats", nats)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), nats); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type ZecNatGatewayFilter struct {
	Ids             []string
	RegionId        string
	Name            string
	ResourceGroupId string
}
