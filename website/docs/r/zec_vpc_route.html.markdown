---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vpc_route"
sidebar_current: "docs-zenlayercloud-resource-zec_vpc_route"
description: |-
  Provides a resource to create a VPC route
---

# zenlayercloud_zec_vpc_route

Provides a resource to create a VPC route

## Example Usage

Prepare VPC, subnet & NIC

```hcl
resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/24"
  enable_ipv6 = true
  mtu         = 1300
}

resource "zenlayercloud_zec_subnet" "subnet" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name      = "example"
}
```

Create vpc route

```hcl
resource "zenlayercloud_zec_vpc_route" "example" {
  vpc_id                 = zenlayercloud_zec_vpc.foo.id
  ip_version             = "IPv4"
  route_type             = "RouteTypeStatic"
  destination_cidr_block = "192.168.0.0/24"
  next_hop_id            = zenlayercloud_zec_vnic.vnic.id
  name                   = "example-route"
  priority               = 10
}
```

## Argument Reference

The following arguments are supported:

* `destination_cidr_block` - (Required, String) Destination address block.
* `ip_version` - (Required, String, ForceNew) IP stack type. Valid values: `IPv4`, `IPv6`.
* `name` - (Required, String) The name of the VPC route. The name should start and end with a number or a letter, containing 2 to 63 characters. Only letters, numbers, - and periods (.) are supported.
* `next_hop_id` - (Required, String) ID of next hop instance. Currently only ID of vNIC is valid.
* `priority` - (Required, Int) Priority of the route entry. Valid value: from `0` to `65535`.
* `route_type` - (Required, String) Route type. Valid values: `RouteTypeStatic`, `RouteTypePolicy`.
* `vpc_id` - (Required, String) ID of the VPC.
* `source_ip` - (Optional, String) The source IP matched. Required when the `route_type` is `RouteTypePolicy`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Creation time of the VPC route.


## Import

VPC route can be imported, e.g.

```
$ terraform import zenlayercloud_zec_vpc_route.example vpc-route-id
```

