---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_subnet"
sidebar_current: "docs-zenlayercloud-resource-zec_subnet"
description: |-
  Provide a resource to create a subnet.
---

# zenlayercloud_zec_subnet

Provide a resource to create a subnet.

~> **NOTE:** If you want to create a subnet with private ipv6 cidr, you must enable ipv6 for VPC.

## Example Usage

Create a VPC

```hcl
resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/24"
  enable_ipv6 = true
  mtu         = 1300
}

variable "region" {
  default = "asia-southeast-1"
}
```

Create a subnet with IPv4 stack

```hcl
resource "zenlayercloud_zec_subnet" "foo" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}
```

Create a subnet with IPv6 stack

```hcl
resource "zenlayercloud_zec_subnet" "foo" {
  vpc_id    = zenlayercloud_zec_vpc.foo.id
  region_id = var.region
  ipv6_type = "Public"
}
```

Create a subnet with IPv6 & IPv4 stack

```hcl
resource "zenlayercloud_zec_subnet" "foo" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  cidr_block = "10.0.0.0/24"
  ipv6_type  = "Public"
}
```

# Import

Subnet instance can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zec_subnet.foo subnet_id
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required, String, ForceNew) ID of the VPC to be associated.
* `cidr_block` - (Optional, String) The ipv4 cidr block. A network address block which should be a subnet of the three internal network segments (10.0.0.0/8, 172.16.0.0/12 and 192.168.0.0/16).
* `ipv6_type` - (Optional, String) The IPv6 type. Valid values: `Public`, `Private`.
* `name` - (Optional, String) The name of the subnet, the default value is 'Terraform-Subnet'.
* `region_id` - (Optional, String, ForceNew) The region that the subnet locates at.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the subnet.
* `ip_stack_type` - Subnet IP stack type. Values: `IPv4`, `IPv6`, `IPv4_IPv6`.
* `ipv6_cidr_block` - The IPv6 network segment.
* `is_default` - Indicates whether it is the default subnet.


