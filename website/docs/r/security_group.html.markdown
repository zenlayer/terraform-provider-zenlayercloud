---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_security_group"
sidebar_current: "docs-zenlayercloud-resource-security_group"
description: |-
  Provides a resource to create security group.
---

# zenlayercloud_security_group

Provides a resource to create security group.

## Example Usage

```hcl
resource "zenlayercloud_security_group" "foo" {
  name        = "example-name"
  description = "example purpose"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The name of the security group.
* `description` - (Optional, String) The name of the security group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Security group can be imported, e.g.

```
$ terraform import zenlayercloud_security_group.security_group security_group_id
```

