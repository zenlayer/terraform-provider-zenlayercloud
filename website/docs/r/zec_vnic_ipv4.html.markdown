---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vnic_ipv4"
sidebar_current: "docs-zenlayercloud-resource-zec_vnic_ipv4"
description: |-
  Provides a resource to assign secondary private ipv4(s) from subnet to vNIC.
---

# zenlayercloud_zec_vnic_ipv4

Provides a resource to assign secondary private ipv4(s) from subnet to vNIC.

~> **NOTE:** The vNIC must contains IPv4 ip stack type

## Example Usage

Prepare a vNIC

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name      = "example-vnic"
}
```
```hcl
resource "zenlayercloud_zec_vnic_ipv4" "foo" {
  vnic_id                    = zenlayercloud_zec_vnic.vnic.id
  secondary_private_ip_count = 3
}
```
```hcl
resource "zenlayercloud_zec_vnic_ipv4" "foo" {
  vnic_id                        = ""
  secondary_private_ip_addresses = ["10.0.0.3", "10.0.0.4"]
}
```

## Argument Reference

The following arguments are supported:

* `vnic_id` - (Required, String, ForceNew) The ID of the vNIC.
* `secondary_private_ip_addresses` - (Optional, Set: [`String`]) Assign specified secondary private ipv4 address. This IP address must be an available IP address within the CIDR block of the subnet to which the vNIC belongs.
* `secondary_private_ip_count` - (Optional, Int) The number of newly-applied private IP addresses.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Disk instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_vnic_ipv4.test vnic-id
```

