/*
Use this data source to query DDoS IP instances.

Example Usage

```hcl
data "zenlayercloud_bmc_ddos_ips" "foo" {
	availability_zone = "SEL-A"
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	"github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/common"
)

func dataSourceZenlayerCloudDdosIps() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudDdosIpsRead,
		Schema: map[string]*schema.Schema{
			"ip_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the DDoS IP to be queried.",
			},
			"availability_zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of zone that the DDoS IPs locates at.",
			},
			"associated_instance_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of instance to bind with DDoS IPs to be queried.",
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
			"ip_status": {
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
			"ip_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of DDoS IP. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID  of the DDoS IP.",
						},
						"availability_zone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of zone that the DDoS IP locates at.",
						},
						"ip_charge_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The charge type of DDoS IP.",
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
						"ip_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the DDoS IP.",
						},
						"instance_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The instance id to bind with the DDoS IP.",
						},
						"instance_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The instance name to bind with the DDoS IP.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the DDoS IP.",
						},
						"expired_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expired time of the DDoS IP.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudDdosIpsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "data_source.zenlayercloud_bmc_ddos_ips.read")()
	//
	bmcService := BmcService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	filter := &DDosIpFilter{}
	if v, ok := d.GetOk("ip_ids"); ok {
		ddosIpIds := v.(*schema.Set).List()
		if len(ddosIpIds) > 0 {
			filter.IpIds = toStringList(ddosIpIds)

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
	if v, ok := d.GetOk("ip_status"); ok {
		filter.DdosIpStatus = common.String(v.(string))
	}

	ddosIpAddress, err := bmcService.DescribeDdosIpAddressesByFilter(filter)
	if err != nil {
		return diag.FromErr(err)
	}
	ddosIpList := make([]map[string]interface{}, 0, len(ddosIpAddress))
	ids := make([]string, 0, len(ddosIpAddress))
	for _, ddos := range ddosIpAddress {
		mapping := map[string]interface{}{
			"instance_id":         ddos.InstanceId,
			"instance_name":       ddos.InstanceName,
			"ip_id":               ddos.DdosIpId,
			"ip_charge_type":      ddos.DdosIpChargeType,
			"public_ip":           ddos.IpAddress,
			"availability_zone":   ddos.ZoneId,
			"resource_group_id":   ddos.ResourceGroupId,
			"resource_group_name": ddos.ResourceGroupName,
			"ip_status":           ddos.DdosIpStatus,
			"create_time":         ddos.CreateTime,
			"expired_time":        ddos.ExpiredTime,
		}
		ddosIpList = append(ddosIpList, mapping)
		ids = append(ids, ddos.InstanceId)
	}
	d.SetId(dataResourceIdHash(ids))
	err = d.Set("ip_list", ddosIpList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), ddosIpList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type DDosIpFilter struct {
	IpIds           []string
	InstanceId      *string
	ZoneId          *string
	Ip              *string
	DdosIpStatus    *string
	ResourceGroupId *string
}
