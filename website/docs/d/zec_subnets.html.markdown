---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_subnets"
sidebar_current: "docs-zenlayercloud-datasource-zec_subnets"
description: |-
  Use this data source to query vpc subnets information.
---

# zenlayercloud_zec_subnets

Use this data source to query vpc subnets information.

## Example Usage

Create subnet resource

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zec_subnet" "subnet" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}
```

Query all subnets

```hcl
data "zenlayercloud_zec_subnets" "all" {
}
```

Query subnets by region id

```hcl
data "zenlayercloud_zec_subnets" "foo" {
  region_id = var.region
}
```

Query subnets by ids

```hcl
data "zenlayercloud_zec_subnets" "foo" {
  ids = [zenlayercloud_zec_subnet.subnet.id]
}
```

Query subnets by name regex

```hcl
data "zenlayercloud_zec_subnets" "foo" {
  name_regex = "^test$"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) ID of the subnets to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the subnet list returned.
* `region_id` - (Optional, String) The region that the subnet locates at.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `result` - An information list of subnets. Each element contains the following attributes:
   * `cidr_block` - The IPv4 network segment.
   * `create_time` - Create time of the subnet.
   * `id` - ID of the subnet.
   * `ip_stack_type` - Subnet IP stack type. Values: `IPv4`, `IPv6`, `IPv4_IPv6`.
   * `ipv6_cidr_block` - The IPv6 network segment.
   * `ipv6_type` - The IPv6 type. Valid values: `Public`, `Private`.
   * `is_default` - Indicates whether it is the default subnet.
   * `name` - Name of the subnet.
   * `region_id` - The region that the subnet locates at.
   * `vpc_id` - ID of the VPC to be associated.


