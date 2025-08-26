/*
Use this data source to query eip instances.

Example Usage

```hcl
data "zenlayercloud_bmc_eips" "foo" {
	availability_zone = "SEL-A"
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	common2 "github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/common"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

func dataSourceZenlayerCloudEips() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudEipRead,
		Schema: map[string]*schema.Schema{
			"eip_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the EIP to be queried.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of zone that the EIPs locates at.",
			},
			"associated_instance_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of instance to bind with EIPs to be queried.",
			},
			"resource_group_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of resource group grouped instances to be queried.",
			},
			"public_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The address of elastic ip to be queried.",
			},
			"eip_status": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The status of elastic ip to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"eip_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of EIP. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"eip_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID  of the EIP.",
						},
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of zone that the EIP locates at.",
						},
						"eip_charge_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The charge type of EIP.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of resource group grouped instances to be queried.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of resource group grouped instances to be queried.",
						},
						"public_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The elastic ip address.",
						},
						"eip_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the EIP.",
						},
						"instance_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The instance id to bind with the EIP.",
						},
						"instance_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The instance name to bind with the EIP.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the EIP.",
						},
						"expired_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expired time of the EIP.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudEipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer common2.LogElapsed(ctx, "data_source.zenlayercloud_bmc_eips.read")()
	//
	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	filter := &EipFilter{}
	if v, ok := d.GetOk("eip_ids"); ok {
		eipIds := v.(*schema.Set).List()
		if len(eipIds) > 0 {
			filter.EipIds = common2.ToStringList(eipIds)

		}
	}
	if v, ok := d.GetOk("availability_zone"); ok {
		filter.ZoneId = common.String(v.(string))
	}
	if v, ok := d.GetOk("associated_instance_id"); ok {
		filter.InstanceId = common.String(v.(string))
	}
	if v, ok := d.GetOk("resource_group_id"); ok {
		filter.ResourceGroupId = common.String(v.(string))
	}
	if v, ok := d.GetOk("public_ip"); ok {
		filter.Ip = common.String(v.(string))
	}
	if v, ok := d.GetOk("eip_status"); ok {
		filter.EipStatus = common.String(v.(string))
	}

	eipAddress, err := bmcService.DescribeEipAddressesByFilter(filter)
	if err != nil {
		return diag.FromErr(err)
	}
	eipList := make([]map[string]interface{}, 0, len(eipAddress))
	ids := make([]string, 0, len(eipAddress))
	for _, eip := range eipAddress {
		mapping := map[string]interface{}{
			"instance_id":         eip.InstanceId,
			"instance_name":       eip.InstanceName,
			"eip_id":              eip.EipId,
			"eip_charge_type":     eip.EipChargeType,
			"public_ip":           eip.IpAddress,
			"availability_zone":   eip.ZoneId,
			"resource_group_id":   eip.ResourceGroupId,
			"resource_group_name": eip.ResourceGroupName,
			"eip_status":          eip.EipStatus,
			"create_time":         eip.CreateTime,
			"expired_time":        eip.ExpiredTime,
		}
		eipList = append(eipList, mapping)
		ids = append(ids, eip.InstanceId)
	}
	d.SetId(common2.DataResourceIdHash(ids))
	err = d.Set("eip_list", eipList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := common2.WriteToFile(output.(string), eipList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type EipFilter struct {
	EipIds          []string
	InstanceId      *string
	ZoneId          *string
	Ip              *string
	EipStatus       *string
	ResourceGroupId *string
}
