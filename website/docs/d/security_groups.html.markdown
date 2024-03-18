---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_security_groups"
sidebar_current: "docs-zenlayercloud-datasource-security_groups"
description: |-
  Use this data source to query detailed information of security groups.
---

# zenlayercloud_security_groups

Use this data source to query detailed information of security groups.

## Example Usage

```hcl
data "zenlayercloud_security_groups" "sg1" {
}

data "zenlayercloud_security_groups" "sg2" {
  name = "example_name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, String) Name of the security group to be queried..
* `result_output_file` - (Optional, String) Used to save results.
* `security_group_id` - (Optional, String) ID of the security group to be queried..

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `security_groups` - An information list of security group. Each element contains the following attributes:
   * `create_time` - Creation time of the security group.
   * `description` - Description of the security group.
   * `instance_ids` - Instance ids of the security group.
   * `name` - Name of the security group.
   * `rule_infos` - Rules set of the security.
      * `cidr_ip` - The cidr ip of the rule.
      * `direction` - The direction of the rule.
      * `ip_protocol` - The protocol of the rule.
      * `policy` - The policy of the rule.
      * `port_range` - The port range of the rule.
   * `security_group_id` - ID of the security group.


