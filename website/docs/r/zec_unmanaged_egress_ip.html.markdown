---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_unmanaged_egress_ip"
sidebar_current: "docs-zenlayercloud-resource-zec_unmanaged_egress_ip"
description: |-
  Provides a resource to manage the rate limit mode of an existing unmanaged egress IP.
---

# zenlayercloud_zec_unmanaged_egress_ip

Provides a resource to manage the rate limit mode of an existing unmanaged egress IP.

This resource does NOT create or delete the unmanaged egress IP itself; it only adopts an existing unmanaged egress IP into Terraform state and allows updates to its `rate_limit_mode`. Destroying this resource only removes it from state.

## Example Usage

```hcl
resource "zenlayercloud_zec_unmanaged_egress_ip" "demo" {
  unmanaged_egress_ip_id = "xxxxxxxxxxxxxxxxx"
  rate_limit_mode        = "LOOSE"
}
```

## Argument Reference

The following arguments are supported:

* `rate_limit_mode` - (Required, String) Bandwidth rate limit mode. Valid values: `LOOSE`, `STRICT`.
* `unmanaged_egress_ip_id` - (Required, String, ForceNew) The ID of the unmanaged egress IP. The IP must already exist; this resource does not create it.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `bandwidth_cap` - Bandwidth cap, measured in Mbps. Null if there is no fixed bandwidth.
* `create_time` - Creation time of the unmanaged egress IP.
* `internet_charge_type` - Internet charge type.
* `ip` - The public IP address.
* `network_line_type` - Network line type.
* `region_id` - The region ID that the unmanaged egress IP locates at.
* `status` - Status of the unmanaged egress IP.
* `vpc_id` - ID of the VPC that the unmanaged egress IP belongs to.


## Import

Unmanaged egress IP rate limit mode can be imported, e.g.

```
$ terraform import zenlayercloud_zec_unmanaged_egress_ip.demo unmanaged-egress-ip-id
```

