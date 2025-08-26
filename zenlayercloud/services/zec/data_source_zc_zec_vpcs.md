Use this data source to query vpc information.

Example Usage

Create a VPC instance using the following steps:
```hcl
resource "zenlayercloud_zec_vpc" "foo" {
  name = "example"
  cidr_block = "10.0.0.0/16"
  enable_ipv6 = true
}
```


Query vpc list by filter
```hcl

data "zenlayercloud_zec_vpcs" "all" {
}

data "zenlayercloud_zec_vpcs" "vpcs" {
  ids = [zenlayercloud_zec_vpc.foo.id]
}

data "zenlayercloud_zec_vpcs" "vpcs" {
  resource_group_id = zenlayercloud_zec_vpc.foo.resource_group_id
}

data "zenlayercloud_zec_vpcs" "vpcs" {
  cidr_block = zenlayercloud_zec_vpc.foo.cidr_block
}

```
