/*
Use this data source to query datacenter ports.

Example Usage

```hcl
data "zenlayercloud_sdn_ports" "foo" {
	datacenter = "SIN1"
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

func dataSourceZenlayerCloudDcPorts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudDcPortsRead,
		Schema: map[string]*schema.Schema{
			"port_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the port to be queried.",
			},
			"datacenter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of datacenter that the port locates at.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"port_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of port. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"port_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the port.",
						},
						"port_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name type of port.",
						},
						"port_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of port. eg. 1G/10G/40G.",
						},
						"remarks": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of port.",
						},
						"datacenter": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The id of datacenter that the port locates at.",
						},
						"datacenter_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of datacenter.",
						},
						"loa_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The LOA state.",
						},
						"loa_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The LOA URL address.",
						},
						"business_entity_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Business entity name. The entity name to be used on the Letter of Authorization (LOA).",
						},
						"port_charge_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The charge type of port.",
						},
						"connect_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The network connectivity state of port.",
						},
						"port_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The business status of port.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the port.",
						},
						"expired_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expired time of the port.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudDcPortsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "data_source.zenlayercloud_sdn_ports.read")()
	//
	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	filter := &PortFilter{}
	if v, ok := d.GetOk("port_ids"); ok {
		portIds := v.(*schema.Set).List()
		if len(portIds) > 0 {
			filter.PortIds = toStringList(portIds)
		}
	}
	if v, ok := d.GetOk("datacenter"); ok {
		filter.DcId = common.String(v.(string))
	}

	ports, err := sdnService.DescribePortsByFilter(filter)
	if err != nil {
		return diag.FromErr(err)
	}
	portList := make([]map[string]interface{}, 0, len(ports))
	ids := make([]string, 0, len(ports))
	for _, port := range ports {
		mapping := map[string]interface{}{
			"port_id":              port.PortId,
			"port_name":            port.PortName,
			"port_type":            port.PortType,
			"remarks":              port.PortRemarks,
			"datacenter":           port.DcId,
			"datacenter_name":      port.DcName,
			"loa_status":           port.LoaStatus,
			"loa_url":              port.LoaDownloadUrl,
			"port_charge_type":     port.PortChargeType,
			"connect_status":       port.ConnectionStatus,
			"business_entity_name": port.BusinessEntityName,
			"port_status":          port.PortStatus,
			"create_time":          port.CreatedTime,
			"expired_time":         port.ExpiredTime,
		}
		portList = append(portList, mapping)
		ids = append(ids, port.PortId)
	}
	d.SetId(dataResourceIdHash(ids))
	err = d.Set("port_list", portList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), portList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

type PortFilter struct {
	PortIds []string
	DcId    *string
}
