package zec

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	zec "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/zec20240401"
	"time"
)

func DataSourceZenlayerCloudZecVpcs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudZecGlobalVpcsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "ID of the global VPC to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group grouped global VPC to be queried.",
			},
			"cidr_block": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter global VPC with this CIDR.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"vpc_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of VPC. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the VPC.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the VPC.",
						},
						"cidr_block": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The network address block of the VPC.",
						},
						"mtu": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum transmission unit. This value cannot be changed.",
						},
						"enable_ipv6": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to enable the private IPv6 network segment.",
						},
						"ipv6_cidr_block": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The private IPv6 network segment after `enable_ipv6` is set to `true`.",
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether it is the default global VPC.",
						},
						"security_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the security group.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group grouped VPC to be queried.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Name of resource group grouped VPC to be queried.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the VPC.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudZecGlobalVpcsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_zec_global_vpcs.read")()

	zecService := ZecService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := &ZecVpcFilter{}

	if v, ok := d.GetOk("ids"); ok {
		vpcIds := v.(*schema.Set).List()
		if len(vpcIds) > 0 {
			request.VpcIds = common2.ToStringList(vpcIds)
		}
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		request.CidrBlock = common.String(v.(string))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common.String(v.(string))
	}

	var vpcs []*zec.VpcInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		vpcs, e = zecService.DescribeVpcsByFilter(ctx, request)
		if e != nil {
			return common2.RetryError(ctx, e, common2.InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	vpcList := make([]map[string]interface{}, 0, len(vpcs))
	ids := make([]string, 0, len(vpcs))
	for _, vpc := range vpcs {
		mapping := map[string]interface{}{
			"id":                  vpc.VpcId,
			"name":                vpc.Name,
			"mtu":                 vpc.Mtu,
			"cidr_block":          vpc.CidrBlock,
			"ipv6_cidr_block":     vpc.Ipv6CidrBlock,
			"is_default":          vpc.IsDefault,
			"enable_ipv6":         vpc.Ipv6CidrBlock == "",
			"resource_group_id":   vpc.ResourceGroup.ResourceGroupId,
			"resource_group_name": vpc.ResourceGroup.ResourceGroupName,
			"security_group_id":   vpc.SecurityGroupId,
			"create_time":         vpc.CreateTime,
		}
		vpcList = append(vpcList, mapping)
		ids = append(ids, vpc.VpcId)
	}

	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("vpc_list", vpcList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), vpcList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type ZecVpcFilter struct {
	VpcIds          []string
	CidrBlock       *string
	vpcRegion       *string
	ResourceGroupId *string
}
