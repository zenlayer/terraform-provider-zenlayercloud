/*
 Use this data source to query vpc subnets information.

Example Usage

```hcl
variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_bmc_subnet" "foo" {
  availability_zone = var.availability_zone
  name       		= "subnet_test"
  cidr_block 		= "10.0.0.0/16"
}

data "zenlayercloud_bmc_subnets" "id_subnets" {
  subnet_id = zenlayercloud_bmc_subnet.foo.id
}

data "zenlayercloud_bmc_subnets" "name_subnets" {
  subnet_name = zenlayercloud_bmc_subnet.foo.name
}

*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	bmc "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/bmc20221120"
	"time"
)

func dataSourceZenlayerCloudVpcSubnets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudVpcSubnetsRead,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the VPC to be queried.",
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the subnet to be queried.",
			},
			"subnet_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the subnet to be queried.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Zone of the subnet to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group grouped subnet to be queried.",
			},
			"cidr_block": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter subnet with this CIDR.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"subnet_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of subnet. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The availability zone of the subnet.",
						},
						"vpc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the VPC.",
						},
						"vpc_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the VPC.",
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the subnet.",
						},
						"subnet_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the subnet.",
						},
						"cidr_block": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A network address block of the subnet.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group grouped subnet to be queried.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Name of resource group grouped subnet to be queried.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation time of the subnet.",
						},
						"subnet_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the subnet.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudVpcSubnetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "data_source.zenlayercloud_bmc_subnets.read")()

	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}

	request := &SubnetFilter{}

	if v, ok := d.GetOk("vpc_id"); ok {
		request.VpcId = v.(string)

	}
	if v, ok := d.GetOk("subnet_id"); ok {
		request.SubnetId = v.(string)
	}

	if v, ok := d.GetOk("subnet_name"); ok {
		request.SubnetName = v.(string)
	}

	if v, ok := d.GetOk("cidr_block"); ok {
		request.CidrBlock = v.(string)
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		request.ZoneId = v.(string)
	}

	if v, ok := d.GetOk("resource_group_id"); ok {
		request.ResourceGroupId = v.(string)
	}

	var subnets []*bmc.Subnet

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead)-time.Minute, func() *resource.RetryError {
		var e error
		subnets, e = bmcService.DescribeSubnets(ctx, request)
		if e != nil {
			return retryError(ctx, e, InternalServerError)
		}
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	subnetList := make([]map[string]interface{}, 0, len(subnets))
	ids := make([]string, 0, len(subnets))
	for _, subnet := range subnets {
		mapping := map[string]interface{}{
			"availability_zone":   subnet.ZoneId,
			"subnet_id":           subnet.SubnetId,
			"subnet_name":         subnet.SubnetName,
			"vpc_id":              subnet.VpcId,
			"vpc_name":            subnet.VpcName,
			"cidr_block":          subnet.CidrBlock,
			"resource_group_id":   subnet.ResourceGroupId,
			"resource_group_name": subnet.ResourceGroupName,
			"create_time":         subnet.CreateTime,
			"subnet_status":       subnet.SubnetStatus,
		}
		subnetList = append(subnetList, mapping)
		ids = append(ids, subnet.SubnetId)
	}

	d.SetId(dataResourceIdHash(ids))
	err = d.Set("subnet_list", subnetList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), subnetList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type SubnetFilter struct {
	VpcId           string
	SubnetId        string
	SubnetName      string
	CidrBlock       string
	ZoneId          string
	ResourceGroupId string
}
