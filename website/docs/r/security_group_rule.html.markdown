---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_security_group_rule"
sidebar_current: "docs-zenlayercloud-resource-security_group_rule"
description: |-
  Provides a resource to create security group rule.
---

# zenlayercloud_security_group_rule

Provides a resource to create security group rule.

~> **NOTE:** Single security rule is hardly ordered, use zenlayercloud_security_group_rule_set instead.

## Example Usage

```hcl
resource "zenlayercloud_security_group" "foo" {
  name        = "example-name"
  description = "example purpose"
}

resource "zenlayercloud_security_group_rule" "bar" {
  security_group_id = zenlayercloud_security_group.foo.id
  direction         = "ingress"
  policy            = "accept"
  cidr_ip           = "10.0.0.0/16"
  ip_protocol       = "tcp"
  port_range        = "80"
}
```

## Argument Reference

The following arguments are supported:

* `cidr_ip` - (Required, String, ForceNew) The cidr ip of the rule.
* `direction` - (Required, String, ForceNew) The direction of the rule.
* `ip_protocol` - (Required, String, ForceNew) The protocol of the rule.
* `port_range` - (Required, String, ForceNew) The port range of the rule.
* `security_group_id` - (Required, String, ForceNew) ID of the security group to be queried.
* `policy` - (Optional, String, ForceNew) The policy of the rule, currently only `accept` is supported.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



