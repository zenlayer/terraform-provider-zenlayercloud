---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_nat_gateway_dnat"
sidebar_current: "docs-zenlayercloud-resource-zec_nat_gateway_dnat"
description: |-
  Provides a resource to create a NAT Gateway DNAT forwarding entry.
---

# zenlayercloud_zec_nat_gateway_dnat

Provides a resource to create a NAT Gateway DNAT forwarding entry.

## Example Usage

Prepare a NAT gateway

```hcl
variable "region_shanghai" {
  default = "asia-east-1"
}


resource "zenlayercloud_zec_nat_gateway" "foo" {
  region_id         = var.region_shanghai
  name              = "test-nat"
  vpc_id            = "<vpc_id>"
  security_group_id = "<security_group_id>"
  subnet_ids        = ["<subnet_id>"]
}
```

Create a DNat forwarding entry

```hcl
resource "zenlayercloud_zec_nat_gateway_dnat" "foo" {
  nat_gateway_id    = zenlayercloud_zec_nat_gateway.foo.id
  source_cidr_block = ["10.0.0.0/8"]
  eip_ids           = ["192.168.1.1"]
}
```

## Argument Reference

The following arguments are supported:

* `eip_id` - (Required, String) ID of the public EIP.
* `nat_gateway_id` - (Required, String, ForceNew) ID of the NAT gateway.
* `private_ip_address` - (Required, String) The private ip address.
* `protocol` - (Required, String) The IP protocol type of the DNAT entry. Valid values: `TCP`, `UDP`, `Any`. If you want to forward all traffic with unchanged ports, please specify the protocol type as `Any` and do not set the internal port and public external port.
* `private_port` - (Optional, String) The internal port or port segment for DNAT rule port forwarding. You can use a hyphen (`-`) to specify a port range, e.g. 80-100. The number of public and private ports must be consistent. The value range is 1-65535.
* `public_port` - (Optional, String) The external public port or port segment for DNAT rule port forwarding. You can use a hyphen (`-`) to specify a port range, e.g. 80-100. The number of public and private ports must be consistent. The value range is 1-65535. If no port is specified, all traffic will be forwarded with the destination port unchanged.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Dnat entry can be imported using the id, the id format must be '{nat_gateway_id}:{dnat_id}'

```
$ terraform import zenlayercloud_zec_nat_gateway_dnat.foo nat-gateway-id:dnat-id
```

