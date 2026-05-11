---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vnic_public_ipv6"
sidebar_current: "docs-zenlayercloud-datasource-zec_vnic_public_ipv6"
description: |-
  Use this data source to query the public IPv6 attached to a single vNIC.
---

# zenlayercloud_zec_vnic_public_ipv6

Use this data source to query the public IPv6 attached to a single vNIC.

## Example Usage

```hcl
data "zenlayercloud_zec_vnic_public_ipv6" "demo" {
  nic_id = "1680855999352675875"
}

output "ipv6_rate_limit_mode" {
  value = data.zenlayercloud_zec_vnic_public_ipv6.demo.rate_limit_mode
}
```

## Argument Reference

The following arguments are supported:

* `nic_id` - (Required, String) ID of the vNIC whose public IPv6 should be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `bandwidth_cluster_id` - ID of the associated shared bandwidth cluster, if any.
* `bandwidth_cluster_name` - Name of the associated shared bandwidth cluster, if any.
* `bandwidth` - Public bandwidth limit of the IPv6, measured in Mbps.
* `internet_charge_type` - Internet charge type of the public IPv6.
* `ipv6_cidr_id` - The IPv6 CIDR ID associated with this public IPv6.
* `ipv6_cidr` - The IPv6 CIDR address.
* `primary_ipv6_address` - The primary IPv6 address of the vNIC.
* `rate_limit_mode` - Bandwidth rate limit mode. `LOOSE` or `STRICT`.
* `traffic_package_size` - Traffic package size of the IPv6, measured in TB.


