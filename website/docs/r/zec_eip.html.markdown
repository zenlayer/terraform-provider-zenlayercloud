---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_eip"
sidebar_current: "docs-zenlayercloud-resource-zec_eip"
description: |-
  Provide a resource to Elastic IP.
---

# zenlayercloud_zec_eip

Provide a resource to Elastic IP.

## Example Usage

Create en EIP billing by flat rate

```hcl
variable "region" {
  default = "asia-southeast-1"
}

resource "zenlayercloud_zec_eip" "eip" {
  region_id            = var.region
  name                 = "example"
  ip_network_type      = "BGPLine"
  internet_charge_type = "ByBandwidth"
  bandwidth            = 10
}
```

## Argument Reference

The following arguments are supported:

* `bandwidth` - (Required, Int) Bandwidth. Measured in Mbps.
* `internet_charge_type` - (Required, String) Network billing methods. Valid values: `ByBandwidth`, `ByTrafficPackage`, `BandwidthCluster`.
* `name` - (Required, String) Name of the elastic IP.
* `region_id` - (Required, String, ForceNew) The region ID that the elastic IP locates at.
* `bandwidth_cluster_id` - (Optional, String) Bandwidth cluster ID. Required when `internet_charge_type` is `BandwidthCluster`.
* `cidr_id` - (Optional, String, ForceNew) CIDR ID, the elastic ip will allocated from given CIDR.
* `flow_package_size` - (Optional, Float64, ForceNew) The Data transfer package. Measured in TB.
* `ip_network_type` - (Optional, String, ForceNew) Network types of public IPv4. Valid values: `CN2Line`, `LocalLine`, `ChinaTelecom`, `ChinaUnicom`, `ChinaMobile`, `Cogent`.
* `peer_region_id` - (Optional, String, ForceNew) Remote region ID.
* `resource_group_id` - (Optional, String) Resource group ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Creation time of the elastic IP.
* `public_ip_address` - The elastic ipv4 address.
* `resource_group_name` - The Name of resource group.
* `status` - Status of the elastic IP.


## Import

EIP instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_eip.eip eip-id
```

