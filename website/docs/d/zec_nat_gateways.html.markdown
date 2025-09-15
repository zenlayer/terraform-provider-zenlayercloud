---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_nat_gateways"
sidebar_current: "docs-zenlayercloud-datasource-zec_nat_gateways"
description: |-
  Use this data source to query information of zec NAT gateways.
---

# zenlayercloud_zec_nat_gateways

Use this data source to query information of zec NAT gateways.

## Example Usage

Query all NAT gateways

```hcl
data "zenlayercloud_zec_nat_gateways" "all" {
}
```

Query NAT gateways by id

```hcl
data "zenlayercloud_zec_nat_gateways" "foo" {
  ids = ["<natGatewayId>"]
}
```

Query NAT gateways by region id

```hcl
data "zenlayercloud_zec_nat_gateways" "nat-gateway-hongkong" {
  region_id = "asia-southeast-1"
}
```

Query NAT gateways by name regex

```hcl
data "zenlayercloud_zec_nat_gateways" "nat-gateway-test" {
  name_regex = "test*"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) ids of the NAT gateway to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the NAT gateway list returned.
* `region_id` - (Optional, String) Region of the NAT gateway to be queried.
* `resource_group_id` - (Optional, String) The ID of resource group grouped NAT gateway to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `nats` - An information list of NAT gateways. Each element contains the following attributes:
   * `create_time` - Create time of the NAT gateway.
   * `is_all_subnets` - Indicates whether all the subnets of region is assigned to NAT gateway.
   * `name` - The name of the NAT gateway.
   * `nat_id` - ID of the NAT gateway.
   * `region_id` - The region that the NAT gateway locates at.
   * `resource_group_id` - The resource group id that the NAT gateway belongs to.
   * `resource_group_name` - The resource group name that the NAT gateway belongs to.
   * `status` - The status of NAT gateway.
   * `subnet_ids` - IDs of the subnets to be associated. if this value not set.
   * `vpc_id` - ID of the VPC to be associated.


