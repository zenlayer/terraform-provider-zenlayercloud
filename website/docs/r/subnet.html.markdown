---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_subnet"
sidebar_current: "docs-zenlayercloud-resource-subnet"
description: |-
  Provide a resource to create a subnet.
---

# zenlayercloud_subnet

Provide a resource to create a subnet.

## Example Usage

```hcl
variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_subnet" "foo" {
  availability_zone = var.availability_zone
  name              = "test-subnet"
  cidr_block        = "10.0.0.0/24"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The ID of zone that the subnet locates at.
* `cidr_block` - (Required, String, ForceNew) A network address block which should be a subnet of the three internal network segments (10.0.0.0/24, 172.16.0.0/24 and 192.168.0.0/24).
* `description` - (Optional, String) The description of subnet.
* `name` - (Optional, String) The name of the subnet, the default value is 'Terraform-Subnet'.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the subnet.
* `subnet_status` - Current status of the subnet.


## Import

Subnet instance can be imported, e.g.

```
$ terraform import zenlayercloud_subnet.subnet subnet_id
```

