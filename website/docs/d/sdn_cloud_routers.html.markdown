---
subcategory: "Cloud Networking(SDN)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_sdn_cloud_routers"
sidebar_current: "docs-zenlayercloud-datasource-sdn_cloud_routers"
description: |-
  Use this data source to query layer 3 cloud routers.
---

# zenlayercloud_sdn_cloud_routers

Use this data source to query layer 3 cloud routers.

## Example Usage

```hcl
data "zenlayercloud_sdn_cloud_routers" "all" {

}
```

## Argument Reference

The following arguments are supported:

* `cr_ids` - (Optional, Set: [`String`]) IDs of the cloud router to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `cr_list` - An information list of cloud router. Each element contains the following attributes:
  * `connectivity_status` - Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.
  * `cr_description` - The description of cloud router.
  * `cr_id` - ID of the cloud router.
  * `cr_name` - The name of cloud router.
  * `cr_status` - The business status of cloud router.
  * `create_time` - Create time of the cloud router.
  * `edge_points` - The access points of cloud router.
    * `bandwidth` - The bandwidth cap of the access point.
    * `bgp_asn` - BGP ASN of the user.
    * `bgp_local_asn` - BGP ASN of the zenlayer. For Tencent, AWS, GOOGLE and Port, this value is 62610.
    * `bpg_password` - BGP key of the user.
    * `cloud_account` - The account of public cloud access point. If cloud type is GOOGLE, the value is google pairing key. This value is available only when point type within cloud type (AWS, GOOGLE and TENCENT).
    * `cloud_region` - Region of cloud access point. This value is available only when point type within cloud type (AWS, GOOGLE and TENCENT).
    * `connectivity_status` - Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.
    * `create_time` - Create time of the access point.
    * `datacenter` - The ID of the datacenter where the access point located.
    * `ip_address` - The interconnect IP address of DC within Zenlayer.
    * `point_id` - The ID of the access point.
    * `point_name` - The name of the access point.
    * `point_type` - The type of the access point, Valid values: (PORT, VPC, AWS, GOOGLE and TENCENT).
    * `port_id` - The ID of the port associated with point. Valid only when port_type is PORT.
    * `static_routes` - Static route.
      * `next_hop` - Next Hop address.
      * `prefix` - The network address to route to nextHop.
    * `vlan_id` - Vlan ID of the access point.  Valid value ranges: [1-4000].
    * `vpc_id` - The ID of the VPC associated with point. Valid only when port_type is VPC.
  * `expired_time` - Expired time of the cloud router.
  * `resource_group_id` - The resource group ID.
  * `resource_group_name` - The Name of resource group.


