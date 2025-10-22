---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_nat_gateway"
sidebar_current: "docs-zenlayercloud-resource-zec_nat_gateway"
description: |-
  Provide a resource to create a NAT gateway.
---

# zenlayercloud_zec_nat_gateway

Provide a resource to create a NAT gateway.

~> **NOTE:** Please use `zenlayercloud_zec_eip_association` to bind Elastic IP (EIP)

## Example Usage

```hcl
variable "region_shanghai" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "vpc" {
  name        = "example"
  cidr_block  = "10.0.0.0/24"
  enable_ipv6 = true
  mtu         = 1300
}

resource "zenlayercloud_zec_subnet" "subnet" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

resource "zenlayercloud_zec_security_group" "sg" {
  name = "example-name"
}

# omit security group rules

resource "zenlayercloud_zec_eip" "eip" {
  region_id            = var.region
  name                 = "example"
  ip_network_type      = "BGPLine"
  internet_charge_type = "ByBandwidth"
  bandwidth            = 10
}

# NAT gateway
resource "zenlayercloud_zec_nat_gateway" "foo" {
  region_id         = var.region_shanghai
  name              = "test-nat"
  vpc_id            = zenlayercloud_zec_vpc.vpc.id
  security_group_id = zenlayercloud_zec_security_group.sg.id
  subnet_ids        = [zenlayercloud_zec_subnet.subnet.id]
}
```

## Argument Reference

The following arguments are supported:

* `region_id` - (Required, String, ForceNew) The region that the NAT gateway locates at.
* `security_group_id` - (Required, String) The ID of a security group.
* `vpc_id` - (Required, String, ForceNew) ID of the VPC to be associated.
* `enable_icmp_reply` - (Optional, Bool) Indicates whether ICMP replay is enabled. Default is disabled.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the NAT gateway. Default is `true`. If set true, the NAT gateway will be permanently deleted instead of being moved into the recycle bin.
* `is_all_subnets` - (Optional, Bool) Indicates whether all the subnets of region is assigned to NAT gateway. This field is conflict with `subnet_ids`.
* `name` - (Optional, String) The name of the NAT gateway, the default value is 'Terraform-Subnet'.
* `resource_group_id` - (Optional, String) The resource group id the NAT gateway belongs to, default to ID of Default Resource Group.
* `subnet_ids` - (Optional, Set: [`String`]) IDs of the subnets to be associated. The subnets must belong to the specified VPC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the NAT gateway.


## Import

NAT Gateway can be imported, e.g.

```
$ terraform import zenlayercloud_zec_nat_gateway.foo nat-gateway-id
```

