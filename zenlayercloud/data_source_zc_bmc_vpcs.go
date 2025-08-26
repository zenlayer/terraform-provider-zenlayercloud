/*
 Use this data source to query vpc information.

Example Usage

```hcl
data "zenlayercloud_bmc_vpc_regions" "region" {
}

resource "zenlayercloud_bmc_vpc" "foo" {
  region     = data.zenlayercloud_bmc_vpc_regions.region.vpc_regions.0.region
  name       = "test_vpc"
  cidr_block = "10.0.0.0/16"
}

data "zenlayercloud_bmc_vpcs" "foo" {
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
	"time"
)

func dataSourceZenlayerCloudVpcs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudVpcsRead,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the VPC to be queried.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "region of the VPC to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group grouped VPC to be queried.",
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: common2.ValidateCIDRNetworkAddress,
				Description:  "Filter VPC with this CIDR.",
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
						"region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The region where the VPC located.",
						},
						"vpc_id": {
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
							Description: "A network address block of the VPC.",
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
						//"subnet_ids": {
						//	Type:     schema.TypeList,
						//	Computed: true,
						//	Elem: &schema.Schema{
						//		Type: schema.TypeString,
						//	},
						//	Description: "A ID list of subnets within this VPC.",
						//},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the VPC.",
						},
						"vpc_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "status of the VPC.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudVpcsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_bmc_vpcs.read")()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := &VpcFilter{}

	if v, ok := d.GetOk("vpc_id"); ok {
		request.VpcId = common.String(v.(string))
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		request.CidrBlock = common.String(v.(string))
	}

	if v, ok := d.GetOk("region"); ok {
		request.vpcRegion = common.String(v.(string))
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = common.String(v.(string))
	}

	var vpcs []*bmc.VpcInfo

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		vpcs, e = bmcService.DescribeVpcsByFilter(ctx, request)
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
			"region":              vpc.VpcRegionId,
			"vpc_id":              vpc.VpcId,
			"name":                vpc.VpcName,
			"cidr_block":          vpc.CidrBlock,
			"resource_group_id":   vpc.ResourceGroupId,
			"resource_group_name": vpc.ResourceGroupName,
			"create_time":         vpc.CreateTime,
			"vpc_status":          vpc.VpcStatus,
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

type VpcFilter struct {
	VpcId           *string
	CidrBlock       *string
	vpcRegion       *string
	ResourceGroupId *string
}
