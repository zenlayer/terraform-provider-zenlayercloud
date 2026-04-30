---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_qos_policy_groups"
sidebar_current: "docs-zenlayercloud-datasource-zec_qos_policy_groups"
description: |-
  Use this data source to query ZEC QoS policy group information.
---

# zenlayercloud_zec_qos_policy_groups

Use this data source to query ZEC QoS policy group information.

## Example Usage

Query all QoS policy groups

```hcl
data "zenlayercloud_zec_qos_policy_groups" "all" {
}
```

Query QoS policy groups by region ID

```hcl
data "zenlayercloud_zec_qos_policy_groups" "by_region" {
  region_id = "asia-southeast-1"
}
```

Query QoS policy groups by IDs

```hcl
data "zenlayercloud_zec_qos_policy_groups" "by_ids" {
  ids = ["<qosPolicyGroupId>"]
}
```

Query QoS policy groups by name regex

```hcl
data "zenlayercloud_zec_qos_policy_groups" "by_name" {
  name_regex = "^example-*"
}
```

Query the QoS policy group that a specific resource belongs to

```hcl
data "zenlayercloud_zec_qos_policy_groups" "by_member" {
  resource_id = "<eipId>"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) IDs of the QoS policy groups to be queried.
* `name_regex` - (Optional, String) A regex string to filter QoS policy groups by name.
* `region_id` - (Optional, String) The region ID to filter QoS policy groups.
* `resource_group_id` - (Optional, String) Resource group ID to filter QoS policy groups.
* `resource_id` - (Optional, String) A member resource ID (EIP, IPv6 or UNMANAGED egress IP console UUID) to filter groups containing this resource.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `result` - A list of QoS policy groups. Each element contains the following attributes:
   * `bandwidth_limit` - The shared bandwidth limit in Mbps.
   * `create_time` - Creation time of the QoS policy group.
   * `id` - ID of the QoS policy group.
   * `member_count` - The number of members in the group.
   * `members` - The member list of the group.
      * `ip_type` - The IP type of the member.
      * `resource_id` - The resource ID of the member.
   * `name` - Name of the QoS policy group.
   * `rate_limit_mode` - The rate limit mode.
   * `region_id` - The region ID of the QoS policy group.
   * `resource_group_id` - Resource group ID.
   * `resource_group_name` - The name of the resource group.
   * `tags` - The tags of the QoS policy group.


