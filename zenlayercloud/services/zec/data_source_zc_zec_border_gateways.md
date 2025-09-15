Use this data source to query zec border gateway information.

Example Usage

Query all border gateways

```hcl
data "zenlayercloud_zec_border_gateways" "all" {
}
```

Query border gateways by id

```hcl
data "zenlayercloud_zec_border_gateways" "foo" {
  ids = ["<borderGatewayId>"]
}
```

Query border gateways by vpc_id

```hcl
data "zenlayercloud_zec_border_gateways" "foo" {
  vpc_id = ["<vpcId>"]
}
```

Query border gateways by region id

```hcl
data "zenlayercloud_zec_border_gateways" "foo" {
  region_id = "asia-east-1"
}
```

Query border gateways by name regex

```hcl
data "zenlayercloud_zec_border_gateways" "foo" {
   name_regex = "shanghai*"
}
```
