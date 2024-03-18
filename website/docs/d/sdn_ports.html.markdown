---
subcategory: "Cloud Networking(SDN)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_sdn_ports"
sidebar_current: "docs-zenlayercloud-datasource-sdn_ports"
description: |-
  Use this data source to query datacenter ports.
---

# zenlayercloud_sdn_ports

Use this data source to query datacenter ports.

## Example Usage

```hcl
data "zenlayercloud_sdn_ports" "foo" {
  datacenter = "SIN1"
}
```

## Argument Reference

The following arguments are supported:

* `datacenter` - (Optional, String) The ID of datacenter that the port locates at.
* `port_ids` - (Optional, Set: [`String`]) IDs of the port to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `port_list` - An information list of port. Each element contains the following attributes:
   * `business_entity_name` - Business entity name. The entity name to be used on the Letter of Authorization (LOA).
   * `connect_status` - The network connectivity state of port.
   * `create_time` - Create time of the port.
   * `datacenter_name` - The name of datacenter.
   * `datacenter` - The id of datacenter that the port locates at.
   * `expired_time` - Expired time of the port.
   * `loa_status` - The LOA state.
   * `loa_url` - The LOA URL address.
   * `port_charge_type` - The charge type of port.
   * `port_id` - ID of the port.
   * `port_name` - The name type of port.
   * `port_status` - The business status of port.
   * `port_type` - The type of port. eg. 1G/10G/40G.
   * `remarks` - The description of port.


