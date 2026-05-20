---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_havips"
sidebar_current: "docs-zenlayercloud-datasource-zec_havips"
description: |-
  Use this data source to query ZEC HaVip (high-availability virtual IP) information.
---

# zenlayercloud_zec_havips

Use this data source to query ZEC HaVip (high-availability virtual IP) information.

## Example Usage

Query all HaVips

```hcl
data "zenlayercloud_zec_havips" "all" {
}
```

Query HaVips by region

```hcl
data "zenlayercloud_zec_havips" "foo" {
  region_id = "asia-east-1"
}
```

Query HaVips by IDs

```hcl
data "zenlayercloud_zec_havips" "foo" {
  ids = ["<haVipId>"]
}
```

Query HaVips by VPC

```hcl
data "zenlayercloud_zec_havips" "foo" {
  vpc_ids = ["<vpcId>"]
}
```

Query HaVips by subnet

```hcl
data "zenlayercloud_zec_havips" "foo" {
  subnet_ids = ["<subnetId>"]
}
```

Query HaVips by private IP address

```hcl
data "zenlayercloud_zec_havips" "foo" {
  ip_addresses = ["10.0.0.100"]
}
```

Query HaVips by bound instance

```hcl
data "zenlayercloud_zec_havips" "foo" {
  instance_ids = ["<instanceId>"]
}
```

Query HaVips by name regex

```hcl
data "zenlayercloud_zec_havips" "foo" {
  name_regex = "example-havip*"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) IDs of the HaVips to be queried.
* `instance_ids` - (Optional, Set: [`String`]) Return HaVips that are bound to the specified instance IDs.
* `ip_addresses` - (Optional, Set: [`String`]) The private IP addresses to filter HaVips.
* `name_regex` - (Optional, String) A regex string to filter HaVips by name.
* `region_id` - (Optional, String) The region ID where the HaVips are located.
* `result_output_file` - (Optional, String) Used to save results.
* `subnet_ids` - (Optional, Set: [`String`]) The subnet IDs to filter HaVips.
* `vpc_ids` - (Optional, Set: [`String`]) The VPC IDs to filter HaVips.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `result` - A list of HaVips. Each element contains the following attributes:
   * `associated_eips` - The list of EIPs associated with the HaVip.
      * `eip_address` - The EIP address.
      * `eip_id` - The ID of the EIP.
   * `associated_instances` - The list of instance IDs associated with the HaVip.
   * `create_time` - The creation time of the HaVip.
   * `id` - The ID of the HaVip.
   * `ip_address` - The private IPv4 address of the HaVip.
   * `master_instance_id` - The current master instance ID. Null when no instance is bound.
   * `name` - The name of the HaVip.
   * `region_id` - The region ID where the HaVip is located.
   * `security_group_id` - The security group ID associated with the HaVip.
   * `subnet_id` - The subnet ID to which the HaVip belongs.
   * `tags` - The tags associated with the HaVip.
   * `vpc_id` - The VPC ID to which the HaVip belongs.


