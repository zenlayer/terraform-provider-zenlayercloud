---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_eips"
sidebar_current: "docs-zenlayercloud-datasource-bmc_eips"
description: |-
  Use this data source to query eip instances.
---

# zenlayercloud_bmc_eips

Use this data source to query eip instances.

## Example Usage

```hcl
data "zenlayercloud_bmc_eips" "foo" {
  availability_zone = "SEL-A"
}
```

## Argument Reference

The following arguments are supported:

* `associated_instance_id` - (Optional, String) The ID of instance to bind with EIPs to be queried.
* `availability_zone` - (Optional, String) The ID of zone that the EIPs locates at.
* `eip_ids` - (Optional, Set: [`String`]) IDs of the EIP to be queried.
* `eip_status` - (Optional, String) The status of elastic ip to be queried.
* `public_ip` - (Optional, String) The address of elastic ip to be queried.
* `resource_group_id` - (Optional, String) The ID of resource group grouped instances to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `eip_list` - An information list of EIP. Each element contains the following attributes:
  * `availability_zone` - The ID of zone that the EIP locates at.
  * `create_time` - Create time of the EIP.
  * `eip_charge_type` - The charge type of EIP.
  * `eip_id` - ID  of the EIP.
  * `eip_status` - Current status of the EIP.
  * `expired_time` - Expired time of the EIP.
  * `instance_id` - The instance id to bind with the EIP.
  * `instance_name` - The instance name to bind with the EIP.
  * `public_ip` - The elastic ip address.
  * `resource_group_id` - The ID of resource group grouped instances to be queried.
  * `resource_group_name` - The name of resource group grouped instances to be queried.


