---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vnics"
sidebar_current: "docs-zenlayercloud-datasource-zec_vnics"
description: |-
  Use this data source to query vNIC information.
---

# zenlayercloud_zec_vnics

Use this data source to query vNIC information.

## Example Usage

Query all vNICs

```hcl
data "zenlayercloud_zec_instances" "foo" {
}
```

Query vNICs by ids

```hcl
data "zenlayercloud_zec_instances" "foo" {
  ids = ["<vnicId>"]
}
```

Query vNICs by region id

```hcl
data "zenlayercloud_zec_instances" "foo" {
  region_id = "asia-southeast-1"
}
```

Query vNICs by name regex

```hcl
data "zenlayercloud_zec_instances" "foo" {
  name_regex = "test*"
}
```

Query vNICs by subnet id

```hcl
data "zenlayercloud_zec_instances" "foo" {
  subnet_id = "<subnetId>"
}
```

Query vNICs by vpc id

```hcl
data "zenlayercloud_zec_instances" "foo" {
  vpc_id = "<vpcId>"
}
```

Query vNICs by associated ZEC instance id

```hcl
data "zenlayercloud_zec_instances" "foo" {
  instance_id = "<instanceId>"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) ID of the vNICs to be queried.
* `instance_id` - (Optional, String) ID of the ZEC instances to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the vNIC list returned.
* `region_id` - (Optional, String) The region that the vNIC locates at.
* `resource_group_id` - (Optional, String) The ID of resource group grouped VPC to be queried.
* `result_output_file` - (Optional, String) Used to save results.
* `subnet_id` - (Optional, String) ID of the subnet to be queried.
* `vpc_id` - (Optional, String) ID of the global VPC to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `vnics` - An information list of vNICs. Each element contains the following attributes:
   * `create_time` - Create time of the vNIC.
   * `id` - ID of the vNIC.
   * `name` - Name of the vNIC.
   * `primary_ipv4_address` - The primary private IPv4 address of the vNIC.
   * `primary_ipv6_address` - The primary IPv6 address of the vNIC.
   * `primary` - Indicates whether the IP is primary.
   * `private_ips` - A set of intranet IPs. including private ipv4 and ipv6.
   * `public_ips` - A set of public IPs. including EIP and public IPv6.
   * `region_id` - The region that the vNIC locates at.
   * `resource_group_id` - The resource group id that the NAT gateway belongs to.
   * `resource_group_name` - The resource group name that the NAT gateway belongs to.
   * `security_group_id` - ID of the security group.


