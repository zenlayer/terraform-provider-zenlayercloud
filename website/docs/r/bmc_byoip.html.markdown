---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_byoip"
sidebar_current: "docs-zenlayercloud-resource-bmc_byoip"
description: |-
  Provide a resource to create a BYOIP (Bring Your Own IP) in BMC.
---

# zenlayercloud_bmc_byoip

Provide a resource to create a BYOIP (Bring Your Own IP) in BMC.

## Example Usage

```hcl
resource "zenlayercloud_bmc_byoip" "foo" {
  ip_type                     = "IPv4"
  cidr                        = "203.0.113.0/24"
  asn                         = 65001
  public_virtual_interface_id = "xxxxxxxx"
}
```

## Argument Reference

The following arguments are supported:

* `asn` - (Required, Int, ForceNew) ASN number of the announced CIDR block.
* `cidr` - (Required, String, ForceNew) The announced IPv4 or IPv6 CIDR block.
* `ip_type` - (Required, String, ForceNew) IP type. Valid values: `IPv4`, `IPv6`.
* `public_virtual_interface_id` - (Required, String, ForceNew) The unique ID of the public virtual interface (public VLAN).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `available_ip_count` - Number of available IPs in the CIDR block.
* `charge_type` - Charge type of the CIDR block.
* `cidr_block_name` - Name of the CIDR block.
* `cidr_block_type` - Type of the CIDR block.
* `create_time` - Creation time of the CIDR block.
* `expire_time` - Expiration time of the CIDR block.
* `gateway` - Gateway address of the CIDR block.
* `resource_group_id` - The resource group ID the CIDR block belongs to.
* `resource_group_name` - The resource group name the CIDR block belongs to.
* `status` - Current status of the CIDR block.
* `zone_id` - The zone ID that the CIDR block locates at.


## Import

BYOIP can be imported using the cidr block ID, e.g.

```
$ terraform import zenlayercloud_bmc_byoip.foo cidr-block-id
```

