---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_disk_snapshot_policies"
sidebar_current: "docs-zenlayercloud-datasource-zec_disk_snapshot_policies"
description: |-
  Use this data source to query Snapshot policies
---

# zenlayercloud_zec_disk_snapshot_policies

Use this data source to query Snapshot policies

## Example Usage

Query all snapshots policies

```hcl
data "zenlayercloud_zec_disk_snapshot_policies" "all" {}
```

Query snapshot policies by id

```hcl
data "zenlayercloud_zec_disk_snapshot_policies" "foo" {
  ids = ["<snapshotPolicyId>"]
}
```

Query snapshots by name regex

```hcl
data "zenlayercloud_zec_disk_snapshot_policies" "foo" {
  name_regex = "^example"
}
```

Query snapshots by availability zone

```hcl
data "zenlayercloud_zec_disk_snapshot_policies" "foo" {
  availability_zone = "asia-east-1a"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) The availability zone of the auto snapshot policy to be queried.
* `ids` - (Optional, Set: [`String`]) IDs of the auto snapshot policy to be queried.
* `name_regex` - (Optional, String) Name of the auto snapshot policy to be queried.
* `resource_group_id` - (Optional, String) The ID of resource group grouped auto snapshot policy to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `auto_snapshot_policies` - An information list of auto snapshot policy. Each element contains the following attributes:
   * `availability_zone` - The availability zone of the auto snapshot policy.
   * `create_time` - Creation time of the auto snapshot policy.
   * `disk_ids` - List of disk IDs associated with this auto snapshot policy.
   * `disk_num` - Number of disks associated with this auto snapshot policy.
   * `hours` - The hours of day when the auto snapshot policy is triggered.
   * `id` - ID of the auto snapshot policy.
   * `name` - Name of the auto snapshot policy.
   * `repeat_week_days` - The days of week when the auto snapshot policy is triggered. Valid values: 1-7.
   * `resource_group_id` - The ID of resource group.
   * `resource_group_name` - The name of resource group.
   * `retention_days` - Retention days of the auto snapshot policy.


