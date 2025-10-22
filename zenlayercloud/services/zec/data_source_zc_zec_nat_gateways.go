package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
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
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "ids of the NAT gateway to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the NAT gateway list returned.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the VPC to be queried.",
			},
			"security_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the security group to be queried.",
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
							Description: "ID of the VPC.",
						},
						"subnet_ids": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "IDs of the subnets.",
						},
						"eip_ids": {
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "IDs of the EIP associated.",
						},
						"security_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the security group associated.",
						},
						"zbg_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of border gateway associated.",
						},
						"is_all_subnets": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether all the subnets of region is assigned to NAT gateway.",
						},
						"icmp_reply_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether ICMP reply is enabled.",
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

	if v, ok := d.GetOk("security_group_id"); ok {
		request.SecurityGroupId = v.(string)
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		request.VpcId = v.(string)
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

	var result []*zec.NatGateway

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
			"icmp_reply_enabled":  natGateway.IcmpReplyEnabled,
			"security_group_id":   natGateway.SecurityGroupId,
			"eip_ids":         natGateway.EipIds,
			"zbg_id":         natGateway.ZbgId,

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
	SecurityGroupId string
	VpcId           string
	Name            string
	ResourceGroupId string
}
