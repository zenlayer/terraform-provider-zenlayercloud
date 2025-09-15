Use this data source to query vpc subnets information.

Example Usage

Create subnet resource
```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name = "example"
  cidr_block = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zec_subnet" "subnet" {
  vpc_id = zenlayercloud_zec_vpc.foo.id
  region_id	 = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}
```

Query all subnets
```hcl
data "zenlayercloud_zec_subnets" "all" {
}
```

Query subnets by region id

```hcl
data "zenlayercloud_zec_subnets" "foo" {
	region_id = var.region
}
```

Query subnets by ids
```hcl
data "zenlayercloud_zec_subnets" "foo" {
  ids = [zenlayercloud_zec_subnet.subnet.id]
}
```

Query subnets by name regex
```hcl
data "zenlayercloud_zec_subnets" "foo" {
  name_regex = "^test$"
}
```
