/*
Use this data source to query layer 2 private connect.

Example Usage

```hcl
data "zenlayercloud_sdn_private_connects" "all" {

}

data "zenlayercloud_sdn_private_connects" "byIds" {
	connect_ids = ["xxxxxxx"]
}
```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	sdn "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/sdn20230830"
)

func dataSourceZenlayerCloudSdnPrivateConnects() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudSdnPrivateConnectsRead,
		Schema: map[string]*schema.Schema{
			"connect_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the private connect to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"connect_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of private connect. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connect_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the private connect.",
						},
						"connect_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name type of private connect.",
						},
						"endpoints": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The endpoint a & endpoint z of private connect.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the port.",
									},
									"endpoint_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the access point.",
									},
									"vlan_id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "VLAN ID of the access point. Value range: from 1 to 4096.",
									},
									"cloud_region": {
										Type:        schema.TypeString,
										Optional:    true,
										ForceNew:    true,
										Description: "Region of cloud access point. This value is available only when `endpoint_type` within cloud type (AWS, GOOGLE and TENCENT).",
									},
									"cloud_account": {
										Type:        schema.TypeString,
										Optional:    true,
										ForceNew:    true,
										Description: "The account of public cloud access point. IF cloud type is GOOGLE, the value is google pairing key. This value is available only when `endpoint_type` within cloud type (AWS, GOOGLE and TENCENT).",
									},
									"endpoint_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The type of the access point, which contains: PORT,AWS,TENCENT and GOOGLE.",
									},
									"datacenter": {
										Type:        schema.TypeString,
										Computed:    true,
										Elem:        &schema.Schema{Type: schema.TypeInt},
										Description: "The ID of data center where the endpoint located.",
									},
									"connectivity_status": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.",
									},
								},
							},
						},
						"connectivity_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.",
						},
						"connect_bandwidth": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum bandwidth cap limit of a private connect.",
						},
						"connect_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The business state of private connect.",
						},
						"resource_group_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The resource group ID.",
						},
						"resource_group_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Name of resource group.",
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the private connect.",
						},
						"expired_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expired time of the private connect.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudSdnPrivateConnectsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "data_source.zenlayercloud_sdn_private_connects.read")()
	//
	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	filter := &PrivateConnectFilter{}
	if v, ok := d.GetOk("connect_ids"); ok {
		portIds := v.(*schema.Set).List()
		if len(portIds) > 0 {
			filter.ConnectIds = toStringList(portIds)
		}
	}

	privateConnects, err := sdnService.DescribePrivateConnectsByFilter(filter)
	if err != nil {
		return diag.FromErr(err)
	}
	connectList := make([]map[string]interface{}, 0, len(privateConnects))
	ids := make([]string, 0, len(privateConnects))
	for _, connect := range privateConnects {
		mapping := map[string]interface{}{
			"connect_id":          connect.PrivateConnectId,
			"connect_name":        connect.PrivateConnectName,
			"connect_bandwidth":   connect.BandwidthMbps,
			"connectivity_status": connect.ConnectivityStatus,
			"resource_group_id":   connect.ResourceGroupId,
			"resource_group_name": connect.ResourceGroupName,
			"connect_status":      connect.PrivateConnectStatus,
			"create_time":         connect.CreateTime,
			"expired_time":        connect.ExpiredTime,
		}

		var res = make([]interface{}, 0, 2)
		res = append(res, mappingConnectEndpoint(&connect.EndpointA))
		res = append(res, mappingConnectEndpoint(&connect.EndpointZ))
		mapping["endpoints"] = res
		connectList = append(connectList, mapping)
		ids = append(ids, connect.PrivateConnectId)
	}
	d.SetId(dataResourceIdHash(ids))
	err = d.Set("connect_list", connectList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), connectList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func mappingConnectEndpoint(endpoint *sdn.PrivateConnectEndpoint) interface{} {
	m := make(map[string]interface{}, 5)
	if endpoint.EndpointType == POINT_TYPE_PORT {
		m["port_id"] = endpoint.EndpointId
	}
	m["cloud_region"] = endpoint.CloudRegionId
	m["cloud_account"] = endpoint.CloudAccountId

	m["endpoint_name"] = endpoint.EndpointName
	m["vlan_id"] = endpoint.VlanId
	m["endpoint_type"] = endpoint.EndpointType
	m["datacenter"] = endpoint.DataCenter.DcId
	m["connectivity_status"] = endpoint.ConnectivityStatus
	return m
}

type PrivateConnectFilter struct {
	ConnectIds []string
}
