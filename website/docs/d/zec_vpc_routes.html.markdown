---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vpc_routes"
sidebar_current: "docs-zenlayercloud-datasource-zec_vpc_routes"
description: |-
  Use this data source to query vpc route entries.
---

# zenlayercloud_zec_vpc_routes

Use this data source to query vpc route entries.

## Example Usage

Query all vpc route entries

```hcl
data "zenlayercloud_zec_vpc_routes" "all" {
}
```

Query vpc route entries by vpc id

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
  vpc_id = "vpc-xxxxxx"
}
```

Query vpc route entries by name regex

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
  name_regex = "^vpc-"
}
```

Query vpc route entries by destination cidr block

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
  destination_cidr_block = "10.0.0.0/16"
}
```

Query vpc route entries by route type

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
  route_type = "RouteTypeStatic"
}
```

Query vpc route entries by ip version

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
  ip_version = "IPv4"
}
```

## Argument Reference

The following arguments are supported:

* `destination_cidr_block` - (Optional, String) Destination address block to be queried.
* `ids` - (Optional, Set: [`String`]) ID of the route to be queried.
* `ip_version` - (Optional, String) IP stack type. Valid values: `IPv4`, `IPv6`.
* `name_regex` - (Optional, String) A regex string to apply to the vNIC list returned.
* `result_output_file` - (Optional, String) Used to save results.
* `route_type` - (Optional, String) Route type to be queried. Valid values: `RouteTypeStatic`(for static route), `RouteTypePolicy`(for policy route), `RouteTypeSubnet`(for subnet route), `RouteTypeNatGw`(for NAT gateway route), `RouteTypeTransit`(for dynamic route).
* `vpc_id` - (Optional, String) ID of the global VPC to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `routes` - An information list of routes. Each element contains the following attributes:
   * `create_time` - Creation time of the VPC route.
   * `destination_cidr_block` - Destination address block.
   * `id` - ID of the route.
   * `ip_version` - IP stack type. Valid values: `IPv4`, `IPv6`.
   * `next_hop_id` - ID of next hop instance.
   * `next_hop_type` - Type of next hop instance. Valid values: `NIC`(for vNIC), `VPC`(for VPC), `NAT`(for NAT gateway), `ZBG`(for border gateway).
   * `priority` - Priority of the route entry. Valid value: from `0` to `65535`.
   * `route_type` - Route type. Valid values: `RouteTypeStatic`(for static route), `RouteTypePolicy`(for policy route), `RouteTypeSubnet`(for subnet route), `RouteTypeNatGw`(for NAT gateway route), `RouteTypeTransit`(for dynamic route).
   * `source_ip` - The source IP matched.


