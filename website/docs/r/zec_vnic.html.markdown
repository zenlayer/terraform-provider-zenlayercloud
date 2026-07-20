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
  tags = {
    "group" = "test"
  }
}
```

Create a vNIC

```hcl
resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name      = "example-vnic"
}
```

## Argument Reference

The following arguments are supported:

* `subnet_id` - (Required, String, ForceNew) The ID of a VPC subnet.
* `ipv6_bandwidth_cluster_id` - (Optional, String) Bandwidth cluster ID for public IPv6. Required when `ipv6_internet_charge_type` is `BandwidthCluster`. Must not be set for other charge types.
* `ipv6_bandwidth` - (Optional, Int) Bandwidth of public IPv6. Measured in Mbps.
* `ipv6_internet_charge_type` - (Optional, String) Network billing methods for public IPv6. Valid values: `ByBandwidth`, `ByTrafficPackage`, `BandwidthCluster`. Changing this field when a public IPv6 is already assigned forces resource recreation.
* `ipv6_traffic_package_size` - (Optional, Float64) Traffic Package size for public IPv6. Measured in TB. Only valid when `ipv6_internet_charge_type` is `ByTrafficPackage`. Range: >= 0.
* `name` - (Optional, String) The name of the vNIC. Maximum length is 63. Default is `Terraform-vNIC`.
* `rate_limit_mode` - (Optional, String) Bandwidth rate limit mode for public IPv6. Valid values: `LOOSE`, `STRICT`. Only takes effect on public IPv6 with fixed bandwidth (`ipv6_internet_charge_type` = `ByBandwidth`).
* `resource_group_id` - (Optional, String) The resource group id the vNIC belongs to, default to ID of Default Resource Group.
* `security_group_id` - (Optional, String) The ID of a security group. If not specified, the security group attached to the VPC will be used. Ensure the VPC has a security group attached via `zenlayercloud_zec_vpc_security_group_attachment` before applying.
* `stack_type` - (Optional, String) The stack type of the vNIC. Valid values: `IPv4`, `IPv6`, `IPv4_IPv6`. Supports in-place update between `IPv4` and `IPv4_IPv6`.
* `tags` - (Optional, Map) The available tags within this vNIC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the vNIC.
* `primary_ipv4` - The primary IPv4 address of the vNIC.
* `primary_ipv6` - The primary IPv6 address of the vNIC.
* `resource_group_name` - The resource group name the vNIC belongs to, default to Default Resource Group.


## Import

vNIC can be imported, e.g.

```
$ terraform import zenlayercloud_zec_vnic.vnic vnic-id
```

