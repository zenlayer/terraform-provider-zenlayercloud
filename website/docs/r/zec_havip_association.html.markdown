---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_havip_association"
sidebar_current: "docs-zenlayercloud-resource-zec_havip_association"
description: |-
  Provides a resource to bind a ZEC instance to a high-availability virtual IP (HaVip).
---

# zenlayercloud_zec_havip_association

Provides a resource to bind a ZEC instance to a high-availability virtual IP (HaVip).

The bound instance's network interface must reside in the same subnet as the HaVip.

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
}

resource "zenlayercloud_zec_instance" "instance" {
  # ... omit instance configuration
}

resource "zenlayercloud_zec_havip_association" "binding" {
  ha_vip_id   = zenlayercloud_zec_havip.havip.id
  instance_id = zenlayercloud_zec_instance.instance.id
}
```

## Argument Reference

The following arguments are supported:

* `ha_vip_id` - (Required, String, ForceNew) The ID of the HaVip.
* `instance_id` - (Required, String, ForceNew) The ID of the instance to associate. The instance's network interface must be in the same subnet as the HaVip.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

HaVip association can be imported using the id (ha_vip_id:instance_id), e.g.

```
$ terraform import zenlayercloud_zec_havip_association.binding havip-id:instance-id
```

