Use this data source to query ZEC HaVip (high-availability virtual IP) information.

Example Usage

Query all HaVips

```hcl
data "zenlayercloud_zec_havips" "all" {
}
```

Query HaVips by region

```hcl
data "zenlayercloud_zec_havips" "foo" {
  region_id = "asia-east-1"
}
```

Query HaVips by IDs

```hcl
data "zenlayercloud_zec_havips" "foo" {
  ids = ["<haVipId>"]
}
```

Query HaVips by VPC

```hcl
data "zenlayercloud_zec_havips" "foo" {
  vpc_ids = ["<vpcId>"]
}
```

Query HaVips by subnet

```hcl
data "zenlayercloud_zec_havips" "foo" {
  subnet_ids = ["<subnetId>"]
}
```

Query HaVips by private IP address

```hcl
data "zenlayercloud_zec_havips" "foo" {
  ip_addresses = ["10.0.0.100"]
}
```

Query HaVips by bound instance

```hcl
data "zenlayercloud_zec_havips" "foo" {
  instance_ids = ["<instanceId>"]
}
```

Query HaVips by name regex

```hcl
data "zenlayercloud_zec_havips" "foo" {
  name_regex = "example-havip*"
}
```
