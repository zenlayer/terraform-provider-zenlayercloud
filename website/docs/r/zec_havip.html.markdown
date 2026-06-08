---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_havip"
sidebar_current: "docs-zenlayercloud-resource-zec_havip"
description: |-
  Provides a resource to create a ZEC high-availability virtual IP (HaVip). For more information, see [High-Availability Virtual IP](https://docs.console.zenlayer.com/welcome/elastic-compute/01-overview).
---

# zenlayercloud_zec_havip

Provides a resource to create a ZEC high-availability virtual IP (HaVip). For more information, see [High-Availability Virtual IP](https://docs.console.zenlayer.com/welcome/elastic-compute/01-overview).

~> **NOTE:** Make sure the target subnet has available private IP addresses. If `ip_address` is omitted, the system will allocate one automatically from the subnet; if specified, it must be an available IP within the subnet's CIDR block.

## Example Usage

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name       = "example"
  cidr_block = "10.0.0.0/16"
}

resource "zenlayercloud_zec_subnet" "subnet" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "example-subnet"
  cidr_block = "10.0.0.0/24"
}

resource "zenlayercloud_zec_havip" "havip" {
  subnet_id = zenlayercloud_zec_subnet.subnet.id
  name      = "example-havip"
  tags = {
    "group" = "test"
  }
}
```

Create HaVip with a specified private IP and security group

```hcl
resource "zenlayercloud_zec_havip" "havip" {
  subnet_id         = zenlayercloud_zec_subnet.subnet.id
  name              = "example-havip"
  ip_address        = "10.0.0.100"
  security_group_id = "sg-xxxxxxxx"
}
```

## Argument Reference

The following arguments are supported:

* `subnet_id` - (Required, String, ForceNew) The ID of the subnet to which the HaVip belongs.
* `ip_address` - (Optional, String, ForceNew) The private IPv4 address of the HaVip. Must be one of the available IPs in the subnet's CIDR block. If not specified, the system will allocate one automatically from the subnet.
* `name` - (Optional, String) The name of the HaVip. Length must be between 1 and 64 characters. Default is `Terraform-HaVip`.
* `security_group_id` - (Optional, String, ForceNew) The ID of the security group. If not specified, the default security group of the VPC will be used.
* `tags` - (Optional, Map) The tags associated with the HaVip.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `associated_eips` - The list of EIPs associated with the HaVip.
   * `eip_address` - The EIP address.
   * `eip_id` - The ID of the EIP.
* `associated_instances` - The list of instance IDs associated with the HaVip.
* `create_time` - The creation time of the HaVip.
* `master_instance_id` - The ID of the current master instance. Null when no instance is bound.
* `region_id` - The region ID where the HaVip is located.
* `vpc_id` - The ID of the VPC to which the HaVip belongs.


## Import

HaVip can be imported using the id, e.g.

```
$ terraform import zenlayercloud_zec_havip.havip havip-id
```

