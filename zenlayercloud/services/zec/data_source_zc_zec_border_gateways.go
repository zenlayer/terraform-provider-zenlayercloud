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

func DataSourceZenlayerCloudBorderGateways() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudBorderGatewaysRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "IDs of the border gateways to be queried.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regex string to apply to the border gateway list returned.",
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "VPC ID of the border gateway to be queried.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Region ID of the border gateway to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},

			"border_gateways": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of border gateways. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zbg_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the border gateway.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the border gateway.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VPC ID that the border gateway belongs to.",
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Region ID of the border gateway.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the border gateway.",
						},
						"asn": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Autonomous System Number.",
						},
						"inter_connect_cidr": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Interconnect IP range.",
						},
						"cloud_router_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Cloud router IDs that border gateway is added into.",
						},
						"advertised_subnet": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Subnet route advertisement.",
						},
						"advertised_cidrs": {
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Custom IPv4 CIDR block list.",
						},
						"nat_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "NAT gateway ID.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudBorderGatewaysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common.LogElapsed(ctx, "data_source.zenlayercloud_zec_border_gateways.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := &BoarderGatewayFilter{}

	if v, ok := d.GetOk("ids"); ok {
		ids := v.(*schema.Set).List()
		idList := make([]string, 0, len(ids))
		for _, id := range ids {
			idList = append(idList, id.(string))
		}
		request.Ids = idList
	}

	var nameRegex *regexp.Regexp
	var errRet error

	if v, ok := d.GetOk("name_regex"); ok {
		nameRegex, errRet = regexp.Compile(v.(string))
		if errRet != nil {
			return diag.Errorf("name_regex format error,%s", errRet.Error())
		}
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		request.VpcId = v.(string)
	}

	if v, ok := d.GetOk("region_id"); ok {
		request.RegionId = v.(string)
	}

	var zbgs []*zec.ZbgInfo
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		zbgs, e = zecService.DescribeBoardGateways(request)
		if e != nil {
			return common.RetryError(ctx, e, common.InternalServerError)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	borderGatewayList := make([]map[string]interface{}, 0, len(zbgs))

	ids := make([]string, 0, len(zbgs))
	for _, borderGateway := range zbgs {
		if nameRegex != nil && !nameRegex.MatchString(borderGateway.Name) {
			continue
		}
		borderGatewayMap := map[string]interface{}{
			"zbg_id":             borderGateway.ZbgId,
			"name":               borderGateway.Name,
			"vpc_id":             borderGateway.VpcId,
			"region_id":          borderGateway.RegionId,
			"create_time":        borderGateway.CreateTime,
			"asn":                borderGateway.Asn,
			"inter_connect_cidr": borderGateway.InterConnectCidr,
			"cloud_router_ids":   borderGateway.CloudRouterIds,
			"advertised_subnet":  borderGateway.AdvertisedSubnet,
			"advertised_cidrs":   borderGateway.AdvertisedCidrs,
			"nat_id":             borderGateway.NatId,
		}

		borderGatewayList = append(borderGatewayList, borderGatewayMap)
		ids = append(ids, borderGateway.ZbgId)
	}

	d.SetId(common.DataResourceIdHash(ids))
	err = d.Set("border_gateways", borderGatewayList)
	if err != nil {
		return diag.FromErr(err)
	}

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {
		if err := common.WriteToFile(output.(string), borderGatewayList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type BoarderGatewayFilter struct {
	Ids      []string
	VpcId    string
	RegionId string
}
