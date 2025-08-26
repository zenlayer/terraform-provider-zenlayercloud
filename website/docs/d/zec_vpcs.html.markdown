---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vpcs"
sidebar_current: "docs-zenlayercloud-datasource-zec_vpcs"
description: |-
  Use this data source to query vpc information.
---

# zenlayercloud_zec_vpcs

Use this data source to query vpc information.

## Example Usage

Create a VPC instance using the following steps:

```hcl
resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}
```

Query vpc list by filter

```hcl
data "zenlayercloud_zec_vpcs" "all" {
}

data "zenlayercloud_zec_vpcs" "vpcs" {
  ids = [zenlayercloud_zec_vpc.foo.id]
}

data "zenlayercloud_zec_vpcs" "vpcs" {
  resource_group_id = zenlayercloud_zec_vpc.foo.resource_group_id
}

data "zenlayercloud_zec_vpcs" "vpcs" {
  cidr_block = zenlayercloud_zec_vpc.foo.cidr_block
}
```

## Argument Reference

The following arguments are supported:

* `cidr_block` - (Optional, String) Filter global VPC with this CIDR.
* `ids` - (Optional, Set: [`String`]) ID of the global VPC to be queried.
* `resource_group_id` - (Optional, String) The ID of resource group grouped global VPC to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `vpc_list` - An information list of VPC. Each element contains the following attributes:
   * `cidr_block` - The network address block of the VPC.
   * `create_time` - Creation time of the VPC.
   * `enable_ipv6` - Whether to enable the private IPv6 network segment.
   * `id` - ID of the VPC.
   * `ipv6_cidr_block` - The private IPv6 network segment after `enable_ipv6` is set to `true`.
   * `is_default` - Indicates whether it is the default global VPC.
   * `mtu` - The maximum transmission unit. This value cannot be changed.
   * `name` - Name of the VPC.
   * `resource_group_id` - The ID of resource group grouped VPC to be queried.
   * `resource_group_name` - The Name of resource group grouped VPC to be queried.
   * `security_group_id` - ID of the security group.


