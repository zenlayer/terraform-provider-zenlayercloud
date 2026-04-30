---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_qos_policy_group_member"
sidebar_current: "docs-zenlayercloud-resource-zec_qos_policy_group_member"
description: |-
  Provide a resource to manage a single member of a QoS policy group.
---

# zenlayercloud_zec_qos_policy_group_member

Provide a resource to manage a single member of a QoS policy group.

Each resource instance represents one IP (Eip, IPv6, or Unmanaged egress IP) added to a group. All attributes are immutable — to change a member's `ip_type` or move it to a different group, destroy and recreate the resource.

## Example Usage

Add an EIP to a QoS policy group

```hcl
variable "region" {
  default = "asia-southeast-1"
}

resource "zenlayercloud_zec_qos_policy_group" "example" {
  region_id       = var.region
  name            = "example-qos-group"
  bandwidth_limit = 100
  rate_limit_mode = "LOOSE"
}

resource "zenlayercloud_zec_eip" "eip" {
  region_id            = var.region
  name                 = "eip-1"
  ip_network_type      = "BGPLine"
  internet_charge_type = "ByBandwidth"
  bandwidth            = 10
}

resource "zenlayercloud_zec_qos_policy_group_member" "member" {
  qos_policy_group_id = zenlayercloud_zec_qos_policy_group.example.id
  resource_id         = zenlayercloud_zec_eip.eip.id
  ip_type             = "Eip"
}
```

Add multiple EIPs to the same group

```hcl
locals {
  eip_ids = [
    zenlayercloud_zec_eip.eip1.id,
    zenlayercloud_zec_eip.eip2.id,
  ]
}

resource "zenlayercloud_zec_qos_policy_group_member" "members" {
  for_each = toset(local.eip_ids)

  qos_policy_group_id = zenlayercloud_zec_qos_policy_group.example.id
  resource_id         = each.value
  ip_type             = "Eip"
}
```

## Argument Reference

The following arguments are supported:

* `ip_type` - (Required, String, ForceNew) The IP type of the member. Valid values: Eip(elastic ip), Ipv6, UnmanagedEgressIp(for unmanaged egress ip).
* `qos_policy_group_id` - (Required, String, ForceNew) The ID of the QoS policy group.
* `resource_id` - (Required, String, ForceNew) The resource ID of the member (EIP, IPv6 or UNMANAGED egress IP console UUID).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

QoS policy group member can be imported using `<qosPolicyGroupId>:<resourceId>`, e.g.

```
$ terraform import zenlayercloud_zec_qos_policy_group_member.member <qosPolicyGroupId>:<resourceId>
```

