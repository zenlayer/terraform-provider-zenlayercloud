---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_nat_gateway_snat"
sidebar_current: "docs-zenlayercloud-resource-zec_nat_gateway_snat"
description: |-
  Provides a resource to create a NAT Gateway SNat entry.
---

# zenlayercloud_zec_nat_gateway_snat

Provides a resource to create a NAT Gateway SNat entry.

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

Create a SNat entry

```hcl
resource "zenlayercloud_zec_nat_gateway_snat" "foo" {
  nat_gateway_id     = zenlayercloud_zec_nat_gateway.foo.id
  source_cidr_blocks = ["10.0.0.0/8"]
  eip_ids            = ["eip_id>"]
}
```

## Argument Reference

The following arguments are supported:

* `nat_gateway_id` - (Required, String, ForceNew) ID of the NAT gateway.
* `eip_ids` - (Optional, Set: [`String`]) IDs of the public EIPs to be associated. This field is conflict with `is_all_eip`. This field is conflict with `is_all_eip`.
* `is_all_eip` - (Optional, Bool) Indicates whether all the EIPs of region is assigned to SNAT entry. This field is conflict with `eip_ids`.
* `source_cidr_blocks` - (Optional, Set: [`String`]) Source CIDR blocks to be associated with the SNAT entry. Cannot be used with `subnet_ids`.
* `subnet_ids` - (Optional, Set: [`String`]) IDs of the subnets to be associated with the SNAT entry. Cannot be used with `source_cidr_blocks`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Snat entry can be imported using the id, the id format must be '{nat_gateway_id}:{snat_id}'

```
$ terraform import zenlayercloud_zec_nat_gateway_snat.foo nat-gateway-id:snat-id
```

