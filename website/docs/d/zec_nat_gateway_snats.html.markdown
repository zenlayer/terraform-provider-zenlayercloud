---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_nat_gateway_snats"
sidebar_current: "docs-zenlayercloud-datasource-zec_nat_gateway_snats"
description: |-
  Use this data source to query information of SNAT of NAT gateway
---

# zenlayercloud_zec_nat_gateway_snats

Use this data source to query information of SNAT of NAT gateway

## Example Usage

Query all SNATS of NAT gateway

```hcl
data "zenlayercloud_zec_nat_gateway_snats" "all" {
  nat_gateway_id = "<natGatewayId>"
}
```

Query SNAT with public EIP

```hcl
data "zenlayercloud_zec_nat_gateway_snats" "foo" {
  nat_gateway_id = "<natGatewayId>"
  eip_id         = "<eipId>"
}
```

Query SNAT with subnet

```hcl
data "zenlayercloud_zec_nat_gateway_snats" "foo" {
  nat_gateway_id = "<natGatewayId>"
  subnet_id      = "<subnetId>"
}
```

## Argument Reference

The following arguments are supported:

* `nat_gateway_id` - (Required, String) ID of the NAT gateway to be queried.
* `eip_id` - (Optional, String) ID of the EIP to be queried.
* `result_output_file` - (Optional, String) Used to save results.
* `subnet_id` - (Optional, String) ID of the subnet to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `snats` - An information list of NAT gateway snat. Each element contains the following attributes:
   * `eip_ids` - IDs of the public EIPs to be associated.
   * `is_all_eip` - Indicates whether all the EIPs of NAT gateway is assigned to SNAT entry.
   * `snat_id` - ID of the NAT gateway.
   * `source_cidr_blocks` - The source cidr block segment.
   * `subnet_ids` - IDs of the subnets to be associated.


