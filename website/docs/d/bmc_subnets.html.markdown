---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_subnets"
sidebar_current: "docs-zenlayercloud-datasource-bmc_subnets"
description: |-
  Use this data source to query vpc subnets information.
---

# zenlayercloud_bmc_subnets

Use this data source to query vpc subnets information.

## Example Usage



## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) Zone of the subnet to be queried.
* `cidr_block` - (Optional, String) Filter subnet with this CIDR.
* `resource_group_id` - (Optional, String) The ID of resource group grouped subnet to be queried.
* `result_output_file` - (Optional, String) Used to save results.
* `subnet_id` - (Optional, String) ID of the subnet to be queried.
* `subnet_name` - (Optional, String) Name of the subnet to be queried.
* `vpc_id` - (Optional, String) ID of the VPC to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `subnet_list` - An information list of subnet. Each element contains the following attributes:
  * `availability_zone` - The availability zone of the subnet.
  * `cidr_block` - A network address block of the subnet.
  * `create_time` - Creation time of the subnet.
  * `resource_group_id` - The ID of resource group grouped subnet to be queried.
  * `resource_group_name` - The Name of resource group grouped subnet to be queried.
  * `subnet_id` - ID of the subnet.
  * `subnet_name` - Name of the subnet.
  * `subnet_status` - Status of the subnet.
  * `vpc_id` - ID of the VPC.
  * `vpc_name` - Name of the VPC.


