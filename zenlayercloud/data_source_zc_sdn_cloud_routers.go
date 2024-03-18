/*
Use this data source to query layer 3 cloud routers.

Example Usage

```hcl
data "zenlayercloud_sdn_cloud_routers" "all" {

}

```
*/
package zenlayercloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/zenlayer/terraform-provider-zenlayercloud/zenlayercloud/connectivity"
	sdn "github.com/zenlayer/zenlayercloud-sdk-go/zenlayercloud/sdn20230830"
)

func dataSourceZenlayerCloudSdnCloudRouters() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceZenlayerCloudSdnCloudRoutersRead,
		Schema: map[string]*schema.Schema{
			"cr_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "IDs of the cloud router to be queried.",
			},
			"result_output_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Used to save results.",
			},
			// Computed value
			"cr_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "An information list of cloud router. Each element contains the following attributes:",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cr_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the cloud router.",
						},
						"cr_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of cloud router.",
						},
						"cr_description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of cloud router.",
						},
						"cr_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The business status of cloud router.",
						},
						"connectivity_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.",
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
						"edge_points": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The access points of cloud router.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"point_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the access point.",
									},
									"point_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the access point.",
									},
									"datacenter": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the datacenter where the access point located.",
									},
									"point_type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The type of the access point, Valid values: (PORT, VPC, AWS, GOOGLE and TENCENT).",
									},
									"vlan_id": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Vlan ID of the access point.  Valid value ranges: [1-4000].",
									},
									"port_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the port associated with point. Valid only when port_type is PORT.",
									},
									"vpc_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the VPC associated with point. Valid only when port_type is VPC.",
									},
									"bandwidth": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The bandwidth cap of the access point.",
									},
									"route_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										Default:      ROUTE_TYPE_BGP,
										ValidateFunc: validation.StringInSlice(ROUTE_TYPES, false),
										Description:  "Type of the route, and available values include BGP and STATIC. The default value is `BGP`.",
									},
									"bgp_asn": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "BGP ASN of the user.",
									},
									"bgp_local_asn": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "BGP ASN of the zenlayer. For Tencent, AWS, GOOGLE and Port, this value is 62610.",
									},
									"connectivity_status": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.",
									},
									"cloud_region": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Region of cloud access point. This value is available only when point type within cloud type (AWS, GOOGLE and TENCENT).",
									},
									"cloud_account": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The account of public cloud access point. If cloud type is GOOGLE, the value is google pairing key. This value is available only when point type within cloud type (AWS, GOOGLE and TENCENT).",
									},
									"ip_address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The interconnect IP address of DC within Zenlayer.",
									},
									"bpg_password": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "BGP key of the user.",
									},
									"static_routes": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Static route.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"prefix": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "The network address to route to nextHop.",
												},
												"next_hop": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Next Hop address.",
												},
											},
										},
									},
									"create_time": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Create time of the access point.",
									},
								},
							},
						},
						"create_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Create time of the cloud router.",
						},
						"expired_time": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expired time of the cloud router.",
						},
					},
				},
			},
		},
	}
}

func dataSourceZenlayerCloudSdnCloudRoutersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	defer logElapsed(ctx, "data_source.zenlayercloud_sdn_private_connects.read")()
	//
	sdnService := SdnService{
		client: meta.(*connectivity.ZenlayerCloudClient),
	}
	filter := &CloudRouterFilter{}
	if v, ok := d.GetOk("cr_ids"); ok {
		portIds := v.(*schema.Set).List()
		if len(portIds) > 0 {
			filter.CloudRouterIds = toStringList(portIds)
		}
	}

	cloudRouters, err := sdnService.DescribeCloudRoutersByFilter(filter)
	if err != nil {
		return diag.FromErr(err)
	}
	crList := make([]map[string]interface{}, 0, len(cloudRouters))
	ids := make([]string, 0, len(cloudRouters))
	for _, cr := range cloudRouters {
		mapping := map[string]interface{}{
			"cr_id":               cr.CloudRouterId,
			"cr_name":             cr.CloudRouterName,
			"cr_description":      cr.CloudRouterDescription,
			"connectivity_status": cr.ConnectivityStatus,
			"resource_group_id":   cr.ResourceGroupId,
			"resource_group_name": cr.ResourceGroupName,
			"cr_status":           cr.CloudRouterStatus,
			"create_time":         cr.CreateTime,
			"expired_time":        cr.ExpiredTime,
			"edge_points":         mappingEdgePoints(cr.EdgePoints),
		}

		crList = append(crList, mapping)
		ids = append(ids, cr.CloudRouterId)
	}
	d.SetId(dataResourceIdHash(ids))
	err = d.Set("cr_list", crList)
	if err != nil {
		return diag.FromErr(err)
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), crList); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func mappingEdgePoints(points []*sdn.CloudRouterEdgePoint) []interface{} {
	var res = make([]interface{}, 0, len(points))
	for _, point := range points {
		m := make(map[string]interface{}, 3)
		m["point_id"] = point.EdgePointId
		m["point_name"] = point.EdgePointName
		m["point_type"] = point.EdgePointType
		m["vlan_id"] = point.VlanId
		m["port_id"] = point.PortId
		m["vpc_id"] = point.VpcId
		m["cloud_region"] = point.CloudRegionId
		m["cloud_account"] = point.CloudAccountId
		m["bandwidth"] = point.BandwidthMbps
		m["connectivity_status"] = point.ConnectivityStatus
		m["ip_address"] = point.IpAddress
		m["datacenter"] = point.DataCenter.DcId
		m["create_time"] = point.CreateTime
		if point.BgpConnection != nil {
			m["route_type"] = ROUTE_TYPE_BGP
		} else {
			m["route_type"] = ROUTE_TYPE_STATIC
		}
		if point.BgpConnection != nil {
			m["bgp_asn"] = point.BgpConnection.PeerAsn
			m["bgp_local_asn"] = point.BgpConnection.LocalAsn
			m["bpg_password"] = point.BgpConnection.Password
		}
		if len(point.StaticRoutes) > 0 {
			m["static_routes"] = mappingStaticRoutes(point.StaticRoutes)
		}
		res = append(res, m)
	}
	return res
}

func mappingStaticRoutes(routes []*sdn.IPRoute) []interface{} {
	var res = make([]interface{}, 0, len(routes))
	for _, route := range routes {
		m := make(map[string]interface{}, 2)
		m["prefix"] = route.Prefix
		m["next_hop"] = route.NextHop
		res = append(res, m)
	}
	return res
}

type CloudRouterFilter struct {
	CloudRouterIds []string
}
