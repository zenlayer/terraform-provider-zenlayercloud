---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vnic"
sidebar_current: "docs-zenlayercloud-resource-zec_vnic"
description: |-
  Provide a resource to create vNIC.
---

# zenlayercloud_zec_vnic

Provide a resource to create vNIC.

## Example Usage

Create VPC & Subnet

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}

# Create subnet (IPv4 IP stack)
resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}
```

Create a vNIC

```hcl
resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name      = "example-vnic"
}
```

# Import

vNIC can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zec_vnic.vnic vnic-id
```

## Argument Reference

The following arguments are supported:

* `subnet_id` - (Required, String, ForceNew) The ID of a VPC subnet.
* `ipv6_bandwidth_cluster_id` - (Optional, String, ForceNew) Bandwidth cluster ID for public IPv6. Required when `ipv6_internet_charge_type` is `BandwidthCluster`.
* `ipv6_bandwidth` - (Optional, Int, ForceNew) Bandwidth of public IPv6. Measured in Mbps.
* `ipv6_internet_charge_type` - (Optional, String, ForceNew) Network billing methods for public IPv6. Valid values: `ByBandwidth`, `ByTrafficPackage`, `BandwidthCluster`.
* `ipv6_traffic_package_size` - (Optional, Float64, ForceNew) Traffic Package size for public IPv6. Measured in TB. Only valid when `ipv6_internet_charge_type` is `ByTrafficPackage`.
* `name` - (Optional, String) The name of the vNIC. maximum length is 63.
* `resource_group_id` - (Optional, String) The resource group id the vNIC belongs to, default to ID of Default Resource Group.
* `security_group_id` - (Optional, String) The ID of a security group. If absent, the security group under VPC will be used.
* `stack_type` - (Optional, String, ForceNew) The stack type of the subnet. Valid values: `IPv4`, `IPv6`, `IPv4_IPv6`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the vNIC.
* `primary_ipv4` - The primary IPv4 address of the vNIC.
* `primary_ipv6` - The primary IPv6 address of the vNIC.
* `resource_group_name` - The resource group name the vNIC belongs to, default to Default Resource Group.


