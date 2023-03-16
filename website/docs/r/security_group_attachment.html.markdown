---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_security_group_attachment"
sidebar_current: "docs-zenlayercloud-resource-security_group_attachment"
description: |-
  Provides a resource to create a security group attachment
---

# zenlayercloud_security_group_attachment

Provides a resource to create a security group attachment

## Example Usage

```hcl
resource "zenlayercloud_security_group_attachment" "foo" {
  security_group_id = "12364246"
  instance_id       = "62343412426423623"
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, String, ForceNew) The id of instance.
* `security_group_id` - (Required, String, ForceNew) The ID of security group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Security group attachment can be imported using the id, e.g.

```
terraform import zenlayercloud_security_group_attachment.security_group_attachment securityGroupId:instanceId
```

