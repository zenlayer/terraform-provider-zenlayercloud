Provide a resource to create a zec global VPC.

Example Usage

Create Vpc
```hcl

resource "zenlayercloud_zec_vpc" "foo" {
  name = "example"
  cidr_block = "10.0.0.0/24"
  enable_ipv6 = true
  mtu = 1300
}

```

# Import

Global Vpc instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_vpc.test vpc-id
```
