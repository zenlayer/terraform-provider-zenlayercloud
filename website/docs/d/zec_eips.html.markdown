---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_eips"
sidebar_current: "docs-zenlayercloud-datasource-zec_eips"
description: |-
  Use this data source to query zec eip information.
---

# zenlayercloud_zec_eips

Use this data source to query zec eip information.

## Example Usage

Query all eips

```hcl
data "zenlayercloud_zec_eips" "all" {
}
```

Query eips by region id

```hcl
data "zenlayercloud_zec_eips" "foo" {
  region_id = "asia-east-1"
}
```

Query eips by ids

```hcl
data "zenlayercloud_zec_eips" "foo" {
  ids = ["<eipId>"]
}
```

Query eips by public ip address

```hcl
data "zenlayercloud_zec_eips" "foo" {
  public_ip_address = "128.0.0.1"
}
```

Query eips by name regex

```hcl
data "zenlayercloud_zec_eips" "foo" {
  name_regex = "nginx-ip*"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) IDs of the EIPs to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the elastic IP list returned.
* `public_ip_address` - (Optional, String) The elastic ipv4 address.
* `region_id` - (Optional, String) The region ID that the elastic IP locates at.
* `resource_group_id` - (Optional, String) Resource group ID.
* `result_output_file` - (Optional, String) Used to save results.
* `status` - (Optional, String) Status of the elastic IP.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `result` - An information list of EIPs. Each element contains the following attributes:
   * `bandwidth_cluster_id` - Bandwidth cluster ID.
   * `bandwidth_cluster_name` - The name of Bandwidth cluster.
   * `bandwidth` - Bandwidth. Measured in Mbps.
   * `cidr_id` - CIDR ID, the elastic ip allocated from.
   * `create_time` - Creation time of the elastic IP.
   * `flow_package_size` - The Data transfer package. Measured in TB.
   * `id` - ID of the EIP.
   * `internet_charge_type` - Network billing methods.
   * `ip_type` - Network types of public IPv4.
   * `name` - Name of the elastic IP.
   * `peer_region_id` - Remote region ID.
   * `public_ip_address` - The elastic ipv4 address.
   * `region_id` - The region ID that the elastic IP locates at.
   * `resource_group_id` - Resource group ID.
   * `resource_group_name` - The Name of resource group.
   * `status` - Status of the elastic IP.


