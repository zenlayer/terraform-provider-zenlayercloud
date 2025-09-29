---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_security_group_rule_set"
sidebar_current: "docs-zenlayercloud-resource-zec_security_group_rule_set"
description: |-
  Provides a resource to manage security group rules.
---

# zenlayercloud_zec_security_group_rule_set

Provides a resource to manage security group rules.

~> **NOTE:** The current resource is used to manage all the rules of one security group, and it is not allowed for the
same security group to use multiple resources to manage them at the same time.

## Example Usage

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

```hcl
$ terraform import zenlayercloud_zec_security_group_rule_set.foo security-group-id
```

## Argument Reference

The following arguments are supported:

* `security_group_id` - (Required, String, ForceNew) ID of the security group.
* `egress` - (Optional, Set) Set of egress rule.
* `ingress` - (Optional, Set) Set of ingress rule.

The `egress` object supports the following:

* `cidr_block` - (Required, String) An IP address network or CIDR segment.
* `policy` - (Required, String) Rule policy of security group. Valid values: `accept` and `deny`.
* `port` - (Required, String) Range of the port. The available value can be a single port, or a port range, or `-1` which means all. E.g. `80`, `80,90`, `80-90` or `all`. Note: If the `Protocol` value is set to `all`, the `Port` value needs to be set to `-1`.
* `protocol` - (Required, String) Type of IP protocol. Valid values: `tcp`, `udp`, `icmp`, `gre`, `icmpv6` and `all`.
* `description` - (Optional, String) Description of the security group rule.
* `priority` - (Optional, Int) Priority of the security group rule. The smaller the value, the higher the priority. Valid values: `1` to `100`. Default is `1`.

The `ingress` object supports the following:

* `cidr_block` - (Required, String) An IP address network or CIDR segment.
* `policy` - (Required, String) Rule policy of security group. Valid values: `accept` and `deny`.
* `port` - (Required, String) Range of the port. The available value can be a single port, or a port range, or `-1` which means all. E.g. `80`, `80,90`, `80-90` or `all`. Note: If the `Protocol` value is set to `all`, the `Port` value needs to be set to `-1`.
* `protocol` - (Required, String) Type of IP protocol. Valid values: `tcp`, `udp`, `icmp`, `gre`, `icmpv6` and `all`.
* `description` - (Optional, String) Description of the security group rule.
* `priority` - (Optional, Int) Priority of the security group rule. The smaller the value, the higher the priority. Valid values: `1` to `100`. Default is `1`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



