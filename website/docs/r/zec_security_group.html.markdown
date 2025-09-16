---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_security_group"
sidebar_current: "docs-zenlayercloud-resource-zec_security_group"
description: |-
  Provides a resource to create ZEC security group.
---

# zenlayercloud_zec_security_group

Provides a resource to create ZEC security group.

## Example Usage

```hcl
resource "zenlayercloud_zec_security_group" "foo" {
  name = "example-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The name of the security group. The length is 1 to 64 characters. Only letters, numbers, - and periods (.) are supported.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Creation time of the security group.


## Import

Security group can be imported, e.g.

```
$ terraform import zenlayercloud_zec_security_group.security_group security-group-id
```

