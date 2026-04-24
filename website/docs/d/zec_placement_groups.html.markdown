---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_placement_groups"
sidebar_current: "docs-zenlayercloud-datasource-zec_placement_groups"
description: |-
  Use this data source to query ZEC placement groups.
---

# zenlayercloud_zec_placement_groups

Use this data source to query ZEC placement groups.

## Example Usage

Query all placement groups

```hcl
data "zenlayercloud_zec_placement_groups" "foo" {
}
```

Query placement groups by ids

```hcl
data "zenlayercloud_zec_placement_groups" "foo" {
  ids = ["<placementGroupId>"]
}
```

Query placement groups by zone

```hcl
data "zenlayercloud_zec_placement_groups" "foo" {
  zone_id = "asia-east-1a"
}
```

Query placement groups by name regex

```hcl
data "zenlayercloud_zec_placement_groups" "foo" {
  name_regex = "example*"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) IDs of the placement groups to be queried.
* `name_regex` - (Optional, String) A regex string to filter results by placement group name.
* `resource_group_id` - (Optional, String) The ID of resource group to filter placement groups.
* `result_output_file` - (Optional, String) Used to save results.
* `zone_id` - (Optional, String) Zone ID to filter placement groups.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `placement_group_list` - An information list of placement groups. Each element contains the following attributes:
   * `affinity` - The affinity level.
   * `constraint_status` - The constraint satisfaction status of the placement group.
   * `create_time` - Creation time of the placement group.
   * `id` - ID of the placement group.
   * `instance_count` - The number of instances in the placement group.
   * `instance_ids` - The list of instance IDs associated with the placement group.
   * `name` - Name of the placement group.
   * `partition_num` - The number of partitions.
   * `resource_group_id` - The resource group ID.
   * `resource_group_name` - The resource group name.
   * `tags` - The tags of the placement group.
   * `zone_id` - Zone ID of the placement group.


