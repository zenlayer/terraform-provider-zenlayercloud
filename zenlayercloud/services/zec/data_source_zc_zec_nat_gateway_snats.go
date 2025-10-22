package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"time"
)

func DataSourceZenlayerCloudZecNatGatewaySnats() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecNatGatewaySnatsRead,

		Schema: map[string]*schema.Schema{
			"nat_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the NAT gateway to be queried.",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the subnet to be queried.",
			},
			"eip_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the EIP to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"snats": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of NAT gateway snat. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"snat_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the NAT gateway.",
						},
						"subnet_ids": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "IDs of the subnets to be associated.",
						},
						"source_cidr_blocks": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "The source cidr block segment.",
						},
						"eip_ids": {
							Type:        schema.TypeSet,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
							Description: "IDs of the public EIPs to be associated.",
						},
						"is_all_eip": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether all the EIPs of NAT gateway is assigned to SNAT entry.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecNatGatewaySnatsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_nat_gateway_snats.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	natGatewayId := d.Get("nat_gateway_id").(string)

	var eipId string
	if v, ok := d.GetOk("eip_id"); ok {
		eipId = v.(string)
	}
	var subnetId string
	if v, ok := d.GetOk("subnet_id"); ok {
		subnetId = v.(string)
	}

	var natGateway *zec.DescribeNatGatewayDetailResponseParams
	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		natGateway, e = zecService.DescribeNatGatewayDetailById(ctx, natGatewayId)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}
	if natGateway == nil {
		return diag.Errorf("resource NAT gateway %s not exist", natGatewayId)
	}

	//补全筛选条件

	snats := make([]map[string]interface{}, 0, len(natGateway.Snats))

	ids := make([]string, 0, len(natGateway.Snats))
	for _, snat := range natGateway.Snats {
		if eipId != "" && !common2.IsContains(snat.EipIds, eipId) {
			continue
		}

		mapping := map[string]interface{}{
			"snat_id":            snat.SnatEntryId,
			"source_cidr_blocks": snat.Cidrs,
			"eip_ids":            snat.EipIds,
			"is_all_eip":         snat.IsAllEip,
		}

		if len(snat.SnatSubnets) > 0 {
			var subnetIds []string
			for _, subnet := range snat.SnatSubnets {
				subnetIds = append(subnetIds, *subnet.SubnetId)
			}
			mapping["subnet_ids"] = subnetIds
		}

		if subnetId != "" && !common2.IsContains(mapping["subnet_ids"], subnetId) {
			continue
		}
		snats = append(snats, mapping)
		ids = append(ids, *snat.SnatEntryId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("snats", snats)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), snats); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
