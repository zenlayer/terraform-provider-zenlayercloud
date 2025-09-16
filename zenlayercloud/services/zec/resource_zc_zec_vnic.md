Provide a resource to create vNIC.

Example Usage

Create VPC & Subnet

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}

# Create subnet (IPv4 IP stack)
resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

```

Create a vNIC

```hcl
resource "zenlayercloud_zec_vnic" "vnic" {
  subnet_id = zenlayercloud_zec_subnet.ipv4.id
  name      = "example-vnic"
}
```

Import

vNIC can be imported, e.g.

```
$ terraform import zenlayercloud_zec_vnic.vnic vnic-id
```
