---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vpc"
sidebar_current: "docs-zenlayercloud-resource-zec_vpc"
description: |-
  Provide a resource to create a zec global VPC.
---

# zenlayercloud_zec_vpc

Provide a resource to create a zec global VPC.

## Example Usage

Create Vpc

```hcl
resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/24"
  enable_ipv6 = true
  mtu         = 1300
}
```

# Import

Global Vpc instance can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zec_vpc.test vpc-id
```

## Argument Reference

The following arguments are supported:

* `cidr_block` - (Required, String) A network address block which should be a subnet of the three internal network segments (10.0.0.0/8, 172.16.0.0/12 and 192.168.0.0/16).
* `enable_ipv6` - (Optional, Bool) Whether to enable the private IPv6 network segment. Once the ipv6 is enabled, disable it will cause the resource to `ForceNew`.
* `mtu` - (Optional, Int, ForceNew) The maximum transmission unit. This value cannot be changed.
* `name` - (Optional, String) The name of the global VPC.
* `resource_group_id` - (Optional, String) The resource group id the global VPC belongs to, default to ID of Default Resource Group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the global VPC.
* `ipv6_cidr_block` - The private IPv6 network segment after `enable_ipv6` is set to `true`.
* `is_default` - Indicates whether it is the default VPC.
* `resource_group_name` - The resource group name the VPC belongs to, default to Default Resource Group.


