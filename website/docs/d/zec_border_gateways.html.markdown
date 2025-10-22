---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_border_gateways"
sidebar_current: "docs-zenlayercloud-datasource-zec_border_gateways"
description: |-
  Use this data source to query zec border gateway information.
---

# zenlayercloud_zec_border_gateways

Use this data source to query zec border gateway information.

## Example Usage

Query all border gateways

```hcl
data "zenlayercloud_zec_border_gateways" "all" {
}
```

Query border gateways by id

```hcl
data "zenlayercloud_zec_border_gateways" "foo" {
  ids = ["<borderGatewayId>"]
}
```

Query border gateways by vpc_id

```hcl
data "zenlayercloud_zec_border_gateways" "foo" {
  vpc_id = ["<vpcId>"]
}
```

Query border gateways by region id

```hcl
data "zenlayercloud_zec_border_gateways" "foo" {
  region_id = "asia-east-1"
}
```

Query border gateways by name regex

```hcl
data "zenlayercloud_zec_border_gateways" "foo" {
  name_regex = "shanghai*"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) IDs of the border gateways to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the border gateway list returned.
* `region_id` - (Optional, String) Region ID of the border gateway to be queried.
* `result_output_file` - (Optional, String) Used to save results.
* `vpc_id` - (Optional, String) VPC ID of the border gateway to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `border_gateways` - An information list of border gateways. Each element contains the following attributes:
   * `advertised_cidrs` - Custom IPv4 CIDR block list.
   * `advertised_route_ids` - IDs of route which are advertised through Border gateway.
   * `advertised_subnet` - Subnet route advertisement.
   * `asn` - Autonomous System Number.
   * `cloud_router_ids` - Cloud router IDs that border gateway is added into.
   * `create_time` - Creation time of the border gateway.
   * `inter_connect_cidr` - Interconnect IP range.
   * `name` - Name of the border gateway.
   * `nat_id` - NAT gateway ID.
   * `region_id` - Region ID of the border gateway.
   * `routing_mode` - Routing mode of border gateway. Valid values: `Regional`, `Global`.
   * `vpc_id` - VPC ID that the border gateway belongs to.
   * `zbg_id` - ID of the border gateway.


