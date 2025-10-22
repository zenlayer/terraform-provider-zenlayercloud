Provide a resource to create a NAT gateway.

~> **NOTE:** Please use `zenlayercloud_zec_eip_association` to bind Elastic IP (EIP)

Example Usage

```hcl

variable "region_shanghai" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "vpc" {
  name = "example"
  cidr_block = "10.0.0.0/24"
  enable_ipv6 = true
  mtu = 1300
}

resource "zenlayercloud_zec_subnet" "subnet" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

resource "zenlayercloud_zec_security_group" "sg" {
  name       	= "example-name"
}

# omit security group rules

resource "zenlayercloud_zec_eip" "eip" {
  region_id = var.region
  name = "example"
  ip_network_type = "BGPLine"
  internet_charge_type = "ByBandwidth"
  bandwidth = 10
}

# NAT gateway
resource "zenlayercloud_zec_nat_gateway" "foo" {
  region_id         = var.region_shanghai
  name              = "test-nat"
  vpc_id            =  zenlayercloud_zec_vpc.vpc.id
  security_group_id =  zenlayercloud_zec_security_group.sg.id
  subnet_ids = [zenlayercloud_zec_subnet.subnet.id]
}

```

Import

NAT Gateway can be imported, e.g.

```
$ terraform import zenlayercloud_zec_nat_gateway.foo nat-gateway-id
```
