---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_subnet"
sidebar_current: "docs-zenlayercloud-resource-bmc_subnet"
description: |-
  Provide a resource to create a VPC subnet.
---

# zenlayercloud_bmc_subnet

Provide a resource to create a VPC subnet.

## Example Usage

```hcl
variable "region" {
  default = "SEL1"
}

variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_bmc_vpc" "foo" {
  region     = var.region
  name       = "test-vpc"
  cidr_block = "10.0.0.0/16"
}

resource "zenlayercloud_bmc_subnet" "subnet_with_vpc" {
  availability_zone = var.availability_zone
  name              = "test-subnet"
  vpc_id            = zenlayercloud_bmc_vpc.foo.id
  cidr_block        = "10.0.10.0/24"
}

resource "zenlayercloud_bmc_subnet" "subnet" {
  availability_zone = var.availability_zone
  name              = "test-subnet"
  cidr_block        = "10.0.10.0/24"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The ID of zone that the bmc subnet locates at.
* `cidr_block` - (Required, String, ForceNew) A network address block which should be a subnet of the three internal network segments (10.0.0.0/16, 172.16.0.0/12 and 192.168.0.0/16).
* `name` - (Optional, String) The name of the bmc subnet.
* `resource_group_id` - (Optional, String) The resource group id the subnet belongs to, default to Default Resource Group.
* `vpc_id` - (Optional, String, ForceNew) ID of the VPC to be associated.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the subnet.
* `resource_group_name` - The resource group name the subnet belongs to, default to Default Resource Group.
* `subnet_status` - Current status of the subnet.
* `vpc_name` - Name of the VPC to be associated.


## Import

Vpc subnet instance can be imported, e.g.

```
$ terraform import zenlayercloud_bmc_subnet.subnet subnet_id
```

