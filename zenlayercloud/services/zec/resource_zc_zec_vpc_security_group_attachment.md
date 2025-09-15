Provides a resource to bind vpc and security group.

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

Attach security group to VPC
```hcl
		
resource "zenlayercloud_zec_vpc_security_group_attachment" "foo" {
  vpc_id 	 		= zenlayercloud_zec_vpc.foo.id
  security_group_id = "<securityGroupId>"
}

```

# Import

VPC instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_vpc_security_group_attachment.test vpc-id:security-group-id
```
