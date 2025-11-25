---
subcategory: "Zenlayer Private DNS(zdns)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zdns_zone_vpc_set_attachment"
sidebar_current: "docs-zenlayercloud-resource-zdns_zone_vpc_set_attachment"
description: |-
  Use this resource to manage a DNS private zone vpc attachment
---

# zenlayercloud_zdns_zone_vpc_set_attachment

Use this resource to manage a DNS private zone vpc attachment

~> **NOTE:** The current resource is used to manage all the vpc of a DNS private zone.

## Example Usage

1. Prepare a DNS Private zone and a VPC

```hcl
resource "zenlayercloud_zdns_zone" "foo" {
  zone_name     = "example.com"
  remark        = "test"
  proxy_pattern = "RECURSION"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/24"
  enable_ipv6 = true
  mtu         = 1300
}
```

2. Bind VPC to DNS Private zone

```hcl
resource "zenlayercloud_zdns_vpc_set_attachment" "foo" {
  zone_id = zenlayercloud_zdns_zone.foo.id
  vpc_ids = [zenlayercloud_zec_vpc.foo.id]
}
```

## Argument Reference

The following arguments are supported:

* `vpc_ids` - (Required, Set: [`String`]) The IDs of the VPCs to be attached to the private zone.
* `zone_id` - (Required, String, ForceNew) The ID of the private zone.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

DNS private zone vpc attachment can be imported, e.g.

```
$ terraform import zenlayercloud_zdns_vpc_set_attachment.foo zone-id
```

