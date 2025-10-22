---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_eip_association"
sidebar_current: "docs-zenlayercloud-resource-zec_eip_association"
description: |-
  Provides an eip resource associated with resource including vNIC, ZLB, NAT Gateway.
---

# zenlayercloud_zec_eip_association

Provides an eip resource associated with resource including vNIC, ZLB, NAT Gateway.

## Example Usage

Bind EIP to vNIC

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

resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name      = "example-vnic"
}

resource "zenlayercloud_zec_eip" "eip" {
  region_id            = var.region
  name                 = "example"
  ip_network_type      = "BGPLine"
  internet_charge_type = "ByBandwidth"
  bandwidth            = 10
}

resource "zenlayercloud_zec_eip_association" "eip_association" {
  eip_id             = zenlayercloud_zec_eip.eip.id
  associated_id      = zenlayercloud_zec_vnic.vnic.id
  associated_type    = "NIC"
  private_ip_address = zenlayercloud_zec_vnic.vnic.primary_ipv4
}
```

Bind EIP to ZLB

```hcl
resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zlb_instance" "zlb" {
  region_id = var.region
  vpc_id    = zenlayercloud_zec_vpc.foo.id
  zlb_name  = "example-5"
}


resource "zenlayercloud_zlb_listener" "tcp_listener" {
  zlb_id               = zenlayercloud_zlb_instance.zlb.id
  listener_name        = "tcp-listener"
  protocol             = "TCP"
  health_check_enabled = true
  port                 = 8080
  scheduler            = "mh"
  kind                 = "FNAT"
  health_check_type    = "TCP"
}

resource "zenlayercloud_zec_eip" "eip" {
  region_id            = var.region
  name                 = "example"
  ip_network_type      = "BGPLine"
  internet_charge_type = "ByBandwidth"
  bandwidth            = 10
}

#Associate EIP with LB
resource "zenlayercloud_zec_eip_association" "eip_association" {
  eip_id          = zenlayercloud_zec_eip.eip.id
  associated_id   = zenlayercloud_zlb_instance.zlb.id
  associated_type = "LB"
}
```

Bind EIP to NAT Gateway

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "vpc" {
  name        = "example"
  cidr_block  = "10.0.0.0/24"
  enable_ipv6 = true
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
resource "zenlayercloud_zec_nat_gateway" "nat" {
  region_id         = var.region_shanghai
  name              = "test-nat"
  vpc_id            = zenlayercloud_zec_vpc.vpc.id
  security_group_id = zenlayercloud_zec_security_group.sg.id
  subnet_ids        = [zenlayercloud_zec_subnet.subnet.id]
}

resource "zenlayercloud_zec_eip_association" "eip_association" {
  eip_id          = zenlayercloud_zec_eip.eip.id
  associated_id   = zenlayercloud_zec_nat_gateway.nat.id
  associated_type = "NAT"
}
```

## Argument Reference

The following arguments are supported:

* `associated_id` - (Required, String, ForceNew) ID of the instance to associate with the EIP.
* `associated_type` - (Required, String, ForceNew) Type of the associated instance. Valid values: LB(Load balancer.), NIC(vNic), NAT(NAT gateway).
* `eip_id` - (Required, String, ForceNew) ID of the elastic IP.
* `bind_type` - (Optional, String, ForceNew) Elastic IP bind type. Effective when the elastic IP is assigned to a vNIC.
* `private_ip_address` - (Optional, String, ForceNew) Private IP address of the instance. Required if associated_type is `Nic`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

EIP association can be imported, e.g.

```
$ terraform import zenlayercloud_zec_eip_association.eip_association eip-id:associated-id:associated-type
```

