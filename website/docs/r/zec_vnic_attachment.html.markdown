---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vnic_attachment"
sidebar_current: "docs-zenlayercloud-resource-zec_vnic_attachment"
description: |-
  Provides a resource to attach vNIC to a ZEC instance.
---

# zenlayercloud_zec_vnic_attachment

Provides a resource to attach vNIC to a ZEC instance.

~> **NOTE:** The QGA of the instance must be installed before using this resource.

## Example Usage

Prepare an instance and a vNIC

```hcl
variable "region" {
  default = "asia-east-1"
}

variable "availability_zone" {
  default = "asia-east-1a"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example-instance-vpc"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zec_vpc_security_group_attachment" "foo" {
  vpc_id            = zenlayercloud_zec_vpc.foo.id
  security_group_id = "<securityGroupId>"
}

# Create subnet (IPv4 IP stack)
resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "example-instance-subnet"
  cidr_block = "10.0.0.0/24"
}

data "zenlayercloud_key_pairs" "all" {
}

data "zenlayercloud_zec_images" "ubuntu" {
  availability_zone = var.availability_zone
  category          = "Ubuntu"
}

# Create an instance
resource "zenlayercloud_zec_instance" "instance" {
  availability_zone = var.availability_zone
  instance_type     = "z2a.cpu.1"
  image_id          = data.zenlayercloud_zec_images.ubuntu.images.0.id
  instance_name     = "Example-Instance"
  key_id            = data.zenlayercloud_key_pairs.all.key_pairs.0.key_id
  subnet_id         = zenlayercloud_zec_subnet.ipv4.id
  system_disk_size  = 20
}

# Create a vNIC
resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name      = "example"
}
```

Attach vNIC to instance

```hcl
resource "zenlayercloud_zec_vnic_attachment" "foo" {
  instance_id = zenlayercloud_zec_instance.instance.id
  vnic_id     = zenlayercloud_zec_vnic.vnic.id
}
```

# Import

vNIC attachment can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zec_vnic_attachment.foo vnic-id : instance-id
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, String, ForceNew) The ID of a ZEC instance.
* `vnic_id` - (Required, String, ForceNew) The ID of the vNIC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



