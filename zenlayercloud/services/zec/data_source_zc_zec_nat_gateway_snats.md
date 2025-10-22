Use this data source to query information of SNAT of NAT gateway

Example Usage

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
