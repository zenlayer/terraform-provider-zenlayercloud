Provides a resource to create a ZEC high-availability virtual IP (HaVip).

~> **NOTE:** Make sure the target subnet has available private IP addresses. If `ip_address` is omitted, the system will allocate one automatically from the subnet; if specified, it must be an available IP within the subnet's CIDR block.

Example Usage

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name       = "example"
  cidr_block = "10.0.0.0/16"
}

resource "zenlayercloud_zec_subnet" "subnet" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "example-subnet"
  cidr_block = "10.0.0.0/24"
}

resource "zenlayercloud_zec_havip" "havip" {
  subnet_id = zenlayercloud_zec_subnet.subnet.id
  name      = "example-havip"
  tags = {
    "group" = "test"
  }
}
```

Create HaVip with a specified private IP and security group

```hcl
resource "zenlayercloud_zec_havip" "havip" {
  subnet_id         = zenlayercloud_zec_subnet.subnet.id
  name              = "example-havip"
  ip_address        = "10.0.0.100"
  security_group_id = "sg-xxxxxxxx"
}
```

Import

HaVip can be imported using the id, e.g.

```
$ terraform import zenlayercloud_zec_havip.havip havip-id
```
