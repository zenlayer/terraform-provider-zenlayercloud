Provides a resource to create ZEC security group.

Example Usage

```hcl

resource "zenlayercloud_zec_security_group" "foo" {
  name       	= "example-name"
}

```

Import

Security group can be imported, e.g.

```
$ terraform import zenlayercloud_zec_security_group.security_group security-group-id
```
