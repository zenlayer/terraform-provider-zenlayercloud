package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20250901"
	"time"
)

func DataSourceZenlayerCloudZecNatGatewayDnats() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecNatGatewayDnatsRead,

		Schema: map[string]*schema.Schema{
			"nat_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the NAT gateway to be queried.",
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"TCP", "UDP", "Any"}, false),
				Description:  "The IP protocol type of the DNAT entry. Valid values: `TCP`, `UDP`, `Any`.",
			},
			"eip_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the public EIP to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"dnats": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of NAT gateway snat. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dnat_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the DNAT entry.",
						},
						"protocol": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The IP protocol type of the DNAT entry. Valid values: `TCP`, `UDP`, `Any`.",
						},
						"private_ip_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private ip address.",
						},
						"private_port": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The internal port or port segment(separated by '-') for DNAT rule port forwarding. The value range is 1-65535.",
						},
						"public_port": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The external public port or port segment(separated by '-') for DNAT rule port forwarding. The value range is 1-65535.",
						},
						"eip_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the public EIP.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecNatGatewayDnatsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_nat_gateway_dnats.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	natGatewayId := d.Get("nat_gateway_id").(string)

	var eipId string
	if v, ok := d.GetOk("eip_id"); ok {
		eipId = v.(string)
	}
	var protocol string
	if v, ok := d.GetOk("protocol"); ok {
		protocol = v.(string)
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

	dnats := make([]map[string]interface{}, 0, len(natGateway.Dnats))

	ids := make([]string, 0, len(natGateway.Dnats))
	for _, dnat := range natGateway.Dnats {
		if eipId != "" && *dnat.EipId != eipId {
			continue
		}

		if protocol != "" && *dnat.Protocol != protocol {
			continue
		}

		mapping := map[string]interface{}{
			"dnat_id":            dnat.DnatEntryId,
			"protocol":           dnat.Protocol,
			"private_ip_address": dnat.PrivateIp,
			"private_port":       dnat.InternalPort,
			"public_port":        dnat.ListenerPort,
			"eip_id":             dnat.EipId,
		}

		dnats = append(dnats, mapping)
		ids = append(ids, *dnat.DnatEntryId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("dnats", dnats)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), dnats); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
