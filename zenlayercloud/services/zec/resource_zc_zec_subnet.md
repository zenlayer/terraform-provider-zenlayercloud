Provide a resource to create a subnet.

~> **NOTE:** If you want to create a subnet with private ipv6 cidr, you must enable ipv6 for VPC.

Example Usage

Create a VPC

```hcl
resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/24"
  enable_ipv6 = true
  mtu         = 1300
}

variable "region" {
  default = "asia-southeast-1"
}
```

Create a subnet with IPv4 stack

```hcl
resource "zenlayercloud_zec_subnet" "foo" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

```

Create a subnet with IPv6 stack

```hcl
resource "zenlayercloud_zec_subnet" "foo" {
  vpc_id    = zenlayercloud_zec_vpc.foo.id
  region_id = var.region
  ipv6_type = "Public"
}

```

Create a subnet with IPv6 & IPv4 stack

```hcl
resource "zenlayercloud_zec_subnet" "foo" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  cidr_block = "10.0.0.0/24"
  ipv6_type  = "Public"
}

```

Import

Subnet instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_subnet.foo subnet_id
```
