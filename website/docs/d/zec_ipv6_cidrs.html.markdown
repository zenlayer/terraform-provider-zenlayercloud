---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_ipv6_cidrs"
sidebar_current: "docs-zenlayercloud-datasource-zec_ipv6_cidrs"
description: |-
  Use this data source to query public IPv6 CIDR blocks.
---

# zenlayercloud_zec_ipv6_cidrs

Use this data source to query public IPv6 CIDR blocks.

## Example Usage

Query all public IPv6 CIDR blocks

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "all" {}
```

Query IPv6 CIDRs by id

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "foo" {
  ids = ["<cidrId>"]
}
```

Query IPv6 CIDRs by name regex

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "foo" {
  name_regex = "^example"
}
```

Query IPv6 CIDRs by region id

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "foo" {
  region_id = "asia-east-1"
}
```

Query BYOIP IPv6 CIDRs by ASN

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "foo" {
  asn = 62210
}
```

## Argument Reference

The following arguments are supported:

* `asn` - (Optional, Int) ASN number to filter the IPv6 CIDR block list. Only valid for `BYOIP` CIDR blocks.
* `cidr_block` - (Optional, String) IPv6 CIDR block address to filter, e.g. `2400:8a00::/28`.
* `ids` - (Optional, Set: [`String`]) IDs of the public IPv6 CIDR block to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the public IPv6 CIDR block list returned.
* `region_id` - (Optional, String) The region ID that the public IPv6 CIDR block locates at.
* `resource_group_id` - (Optional, String) Resource group ID.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `cidrs` - An information list of IPv6 CIDR blocks. Each element contains the following attributes:
   * `asn` - ASN number. Only meaningful when the IPv6 CIDR block source is `BYOIP`; returns `0` for non-BYOIP CIDR blocks (the underlying API returns null in that case, which Terraform renders as `0` due to the limitation that `TypeInt` cannot represent null).
   * `cidr_block_address` - IPv6 CIDR block address.
   * `create_time` - Creation time of the IPv6 CIDR block.
   * `expired_time` - Expiration time of the IPv6 CIDR block.
   * `id` - ID of the IPv6 CIDR block.
   * `name` - Name of the IPv6 CIDR block.
   * `netmask` - The IPv6 CIDR block size.
   * `network_type` - Network types of the IPv6 CIDR block.
   * `nic_ids` - vNIC IDs that the IPv6 CIDR block is associated with.
   * `region_id` - The region ID that the IPv6 CIDR block locates at.
   * `resource_group_id` - Resource group ID.
   * `resource_group_name` - The Name of resource group.
   * `status` - Status of the IPv6 CIDR block.
   * `subnet_ids` - Subnet IDs that the IPv6 CIDR block is associated with.
   * `tags` - The available tags within this IPv6 CIDR block.
   * `type` - The type of IPv6 CIDR block. Valid values: `Console`(for normal public CIDR), `BYOIP`(for bring your own IP).


