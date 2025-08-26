---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vpc_security_group_attachment"
sidebar_current: "docs-zenlayercloud-resource-zec_vpc_security_group_attachment"
description: |-
  Provides a resource to bind vpc and security group.
---

# zenlayercloud_zec_vpc_security_group_attachment

Provides a resource to bind vpc and security group.

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

Attach security group to VPC

```hcl
resource "zenlayercloud_zec_vpc_security_group_attachment" "foo" {
  vpc_id            = zenlayercloud_zec_vpc.foo.id
  security_group_id = "<securityGroupId>"
}
```

# Import

VPC instance can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zec_vpc_security_group_attachment.test vpc-id : security-group-id
```

## Argument Reference

The following arguments are supported:

* `security_group_id` - (Required, String, ForceNew) The ID of the security group.
* `vpc_id` - (Required, String, ForceNew) The ID of the VPC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



