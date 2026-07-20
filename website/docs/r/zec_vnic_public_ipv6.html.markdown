---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vnic_public_ipv6"
sidebar_current: "docs-zenlayercloud-resource-zec_vnic_public_ipv6"
description: |-
  ~> **Deprecated** Use `zenlayercloud_zec_vnic` with `stack_type`, `ipv6_internet_charge_type`, `ipv6_bandwidth`, and `rate_limit_mode` instead.
---

# zenlayercloud_zec_vnic_public_ipv6

~> **Deprecated** Use `zenlayercloud_zec_vnic` with `stack_type`, `ipv6_internet_charge_type`, `ipv6_bandwidth`, and `rate_limit_mode` instead.

Provides a resource to manage the rate limit mode of an existing public IPv6 address on a vNIC.

This resource does NOT create or delete the public IPv6 address itself; it only adopts an existing public IPv6 (already attached to a vNIC) into Terraform state and allows updates to its `rate_limit_mode`. Destroying this resource only removes it from state.

## Example Usage

```hcl
resource "zenlayercloud_zec_vnic_public_ipv6" "demo" {
  nic_id          = "xxxxxxxxxxxxxxxxx"
  rate_limit_mode = "LOOSE"
}
```

## Argument Reference

The following arguments are supported:

* `nic_id` - (Required, String, ForceNew) The ID of the vNIC. The vNIC must already have a public IPv6 attached; this resource does not create it.
* `rate_limit_mode` - (Required, String) Bandwidth rate limit mode. Valid values: `LOOSE`, `STRICT`. Only takes effect on public IPv6 with a fixed bandwidth.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `bandwidth` - Public bandwidth limit of the IPv6, measured in Mbps.
* `internet_charge_type` - Internet charge type of the public IPv6.
* `ipv6_cidr_id` - The IPv6 CIDR ID associated with the public IPv6.
* `ipv6_cidr` - The IPv6 CIDR address.
* `primary_ipv6_address` - The primary IPv6 address of the vNIC.
* `traffic_package_size` - Traffic package size of the IPv6, measured in TB.


## Import

vNIC public IPv6 rate limit mode can be imported using the vNIC ID, e.g.

```
$ terraform import zenlayercloud_zec_vnic_public_ipv6.demo nic-id
```

