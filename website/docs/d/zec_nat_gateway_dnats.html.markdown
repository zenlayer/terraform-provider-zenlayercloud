---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_nat_gateway_dnats"
sidebar_current: "docs-zenlayercloud-datasource-zec_nat_gateway_dnats"
description: |-
  Use this data source to query information of Dnat forwarding rules of NAT gateway
---

# zenlayercloud_zec_nat_gateway_dnats

Use this data source to query information of Dnat forwarding rules of NAT gateway

## Example Usage

Query all Dnat of NAT gateway

```hcl
data "zenlayercloud_zec_nat_gateway_dnats" "all" {
  nat_gateway_id = "<natGatewayId>"
}
```

Query Dnat with public EIP

```hcl
data "zenlayercloud_zec_nat_gateway_dnats" "foo" {
  nat_gateway_id = "<natGatewayId>"
  eip_id         = "<eipId>"
}
```

Query Dnat by protocol

```hcl
data "zenlayercloud_zec_nat_gateway_dnats" "foo" {
  nat_gateway_id = "<natGatewayId>"
  protocol       = "TCP"
}
```

## Argument Reference

The following arguments are supported:

* `nat_gateway_id` - (Required, String) ID of the NAT gateway to be queried.
* `eip_id` - (Optional, String) ID of the public EIP to be queried.
* `protocol` - (Optional, String) The IP protocol type of the DNAT entry. Valid values: `TCP`, `UDP`, `Any`.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `dnats` - An information list of NAT gateway snat. Each element contains the following attributes:
   * `dnat_id` - ID of the DNAT entry.
   * `eip_id` - ID of the public EIP.
   * `private_ip_address` - The private ip address.
   * `private_port` - The internal port or port segment(separated by '-') for DNAT rule port forwarding. The value range is 1-65535.
   * `protocol` - The IP protocol type of the DNAT entry. Valid values: `TCP`, `UDP`, `Any`.
   * `public_port` - The external public port or port segment(separated by '-') for DNAT rule port forwarding. The value range is 1-65535.


