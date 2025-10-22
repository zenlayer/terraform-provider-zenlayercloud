Use this data source to query information of Dnat forwarding rules of NAT gateway

Example Usage

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
  protocol      = "TCP"
}
```
