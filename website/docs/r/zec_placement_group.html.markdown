---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_placement_group"
sidebar_current: "docs-zenlayercloud-resource-zec_placement_group"
description: |-
  Provides a ZEC placement group resource.
---

# zenlayercloud_zec_placement_group

Provides a ZEC placement group resource.

## Example Usage

```hcl
# Create a placement group
resource "zenlayercloud_zec_placement_group" "foo" {
  zone_id       = "asia-east-1a"
  name          = "example-placement-group"
  partition_num = 3
  affinity      = 1
  tags = {
    "testKey" = "testValue"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, String) The name of the placement group. Must be 2-63 characters, starting and ending with a letter or digit.
* `zone_id` - (Required, String, ForceNew) The zone ID of the placement group. such as 'asia-east-1a'.
* `affinity` - (Optional, Int) The affinity level. Range: 1 to partition_num/2. Default: partition_num/2.
* `partition_num` - (Optional, Int) The number of partitions. Range: 2-5, default: 3.
* `resource_group_id` - (Optional, String) The resource group ID the placement group belongs to.
* `tags` - (Optional, Map) The tags of the placement group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `constraint_status` - The constraint satisfaction status of the placement group.
* `create_time` - The creation time of the placement group.
* `instance_count` - The number of instances in the placement group.
* `instance_ids` - The list of instance IDs associated with the placement group.
* `resource_group_name` - The resource group name the placement group belongs to.


## Import

Placement group can be imported using the id, e.g.

```
terraform import zenlayercloud_zec_placement_group.foo placement-group-id
```

