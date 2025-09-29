---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_cidrs"
sidebar_current: "docs-zenlayercloud-datasource-zec_cidrs"
description: |-
  Use this data source to query public CIDR blocks.
---

# zenlayercloud_zec_cidrs

Use this data source to query public CIDR blocks.

## Example Usage

Query all public CIDR blocks

```hcl
data "zenlayercloud_zec_cidrs" "all" {}
```

Query CIDRs by id

```hcl
data "zenlayercloud_zec_cidrs" "snapshot" {
  ids = ["<cidrId>"]
}
```

Query CIDRs by name regex

```hcl
data "zenlayercloud_zec_cidrs" "foo" {
  name_regex = "^example"
}
```

Query CIDRs by region id

```hcl
data "zenlayercloud_zec_cidrs" "foo" {
  region_id = "asia-east-1"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) IDs of the public CIDR block to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the public CIDR block list returned.
* `region_id` - (Optional, String) The region ID that the public CIDR block locates at.
* `resource_group_id` - (Optional, String) Resource group ID.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `cidrs` - An information list of CIDR blocks. Each element contains the following attributes:
   * `cidr_block_address` - CIDR block address.
   * `create_time` - Creation time of the elastic IP.
   * `id` - ID of the public CIDR block.
   * `name` - Name of the public CIDR block.
   * `netmask` - The IDR block size.
   * `network_type` - Network types of public CIDR block. Valid values: `CN2Line`, `LocalLine`, `ChinaTelecom`, `ChinaUnicom`, `ChinaMobile`, `Cogent`.
   * `region_id` - The region ID that the public CIDR block locates at.
   * `resource_group_id` - Resource group ID.
   * `resource_group_name` - The Name of resource group.
   * `status` - Status of the elastic IP.
   * `type` - The type of CIDR block. Valid values: `Console`(for normal public CIDR), `BYOIP`(for bring your own IP).
   * `used_ip_num` - Quantity of used CIDR IPs.


