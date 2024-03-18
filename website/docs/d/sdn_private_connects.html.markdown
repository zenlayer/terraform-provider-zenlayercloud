---
subcategory: "Cloud Networking(SDN)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_sdn_private_connects"
sidebar_current: "docs-zenlayercloud-datasource-sdn_private_connects"
description: |-
  Use this data source to query layer 2 private connect.
---

# zenlayercloud_sdn_private_connects

Use this data source to query layer 2 private connect.

## Example Usage

```hcl
data "zenlayercloud_sdn_private_connects" "all" {

}

data "zenlayercloud_sdn_private_connects" "byIds" {
  connect_ids = ["xxxxxxx"]
}
```

## Argument Reference

The following arguments are supported:

* `connect_ids` - (Optional, Set: [`String`]) IDs of the private connect to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `connect_list` - An information list of private connect. Each element contains the following attributes:
  * `connect_bandwidth` - Maximum bandwidth cap limit of a private connect.
  * `connect_id` - ID of the private connect.
  * `connect_name` - The name type of private connect.
  * `connect_status` - The business state of private connect.
  * `connectivity_status` - Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.
  * `create_time` - Create time of the private connect.
  * `endpoints` - The endpoint a & endpoint z of private connect.
    * `connectivity_status` - Network connectivity state. ACTIVE means the network is connected. DOWN which means not connected.
    * `datacenter` - The ID of data center where the endpoint located.
    * `endpoint_name` - The name of the access point.
    * `endpoint_type` - The type of the access point, which contains: PORT,AWS,TENCENT and GOOGLE.
    * `port_id` - The ID of the port.
    * `vlan_id` - VLAN ID of the access point. Value range: from 1 to 4096.
  * `expired_time` - Expired time of the private connect.
  * `resource_group_id` - The resource group ID.
  * `resource_group_name` - The Name of resource group.


