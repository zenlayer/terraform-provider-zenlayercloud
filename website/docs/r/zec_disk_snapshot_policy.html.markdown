---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_disk_snapshot_policy"
sidebar_current: "docs-zenlayercloud-resource-zec_disk_snapshot_policy"
description: |-
  Provides a resource to create auto snapshot policy
---

# zenlayercloud_zec_disk_snapshot_policy

Provides a resource to create auto snapshot policy

## Example Usage

```hcl
resource "zenlayercloud_zec_disk_snapshot_policy" "example" {
  availability_zone = "asia-east-1a"
  name              = "example-snapshot-policy"
  repeat_week_days  = [1]
  hours             = [12]
  retention_days    = 7
}
```

, e.g.

```hcl
bash
$ terraform import zc_zec_disk_snapshot_policy.example policy-id
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The availability zone of snapshot policy.
* `hours` - (Required, Set: [`Int`]) The hours of day when the auto snapshot policy is triggered. The time zone of hour is `UTC+0`. Valid values: from `0` to `23`.
* `repeat_week_days` - (Required, Set: [`Int`]) The days of week when the auto snapshot policy is triggered. Valid values: `1` to `7`. 1: Monday, 2: Tuesday ~ 7: Sunday.
* `name` - (Optional, String) The name of the snapshot policy. The name should start and end with a number or a letter, containing 2 to 63 characters. Only letters, numbers, - and periods (.) are supported.
* `resource_group_id` - (Optional, String) The ID of resource group grouped snapshot policy.
* `retention_days` - (Optional, Int) The retention days of the auto snapshot policy. Valid values: `1` to `65535` or `-1` for no expired. Default is `-1`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Creation time of the snapshot policy.
* `resource_group_name` - The Name of resource group grouped snapshot policy.


