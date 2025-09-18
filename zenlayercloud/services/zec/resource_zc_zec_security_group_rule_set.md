Provides a resource to manage security group rules.

~> **NOTE:** The current resource is used to manage all the rules of one security group, and it is not allowed for the
same security group to use multiple resources to manage them at the same time.

Example Usage

```hcl

resource "zenlayercloud_zec_security_group" "foo" {
  name = "example-name"
}

resource "zenlayercloud_zec_security_group_rule_set" "foo" {
  security_group_id = zenlayercloud_zec_security_group.foo.id
  ingress {
    policy     = "accept"
    cidr_block = "0.0.0.0/0"
    protocol   = "tcp"
    port       = "8080"
    priority   = 1
  }
}

```
# Import
Security group rules can be imported, e.g.
```
$ terraform import zenlayercloud_zec_security_group_rule_set.foo security-group-id
 ```
