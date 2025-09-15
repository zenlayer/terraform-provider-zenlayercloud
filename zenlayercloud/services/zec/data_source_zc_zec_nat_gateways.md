Use this data source to query information of zec NAT gateways.

Example Usage

Query all NAT gateways
```hcl
data "zenlayercloud_zec_nat_gateways" "all" {
}
```

Query NAT gateways by id
```hcl
data "zenlayercloud_zec_nat_gateways" "foo" {
  ids = ["<natGatewayId>"]
}
```

Query NAT gateways by region id
```hcl
data "zenlayercloud_zec_nat_gateways" "nat-gateway-hongkong" {
  region_id = "asia-southeast-1"
}
```

Query NAT gateways by name regex
```hcl
data "zenlayercloud_zec_nat_gateways" "nat-gateway-test" {
  name_regex = "test*"
}
```
