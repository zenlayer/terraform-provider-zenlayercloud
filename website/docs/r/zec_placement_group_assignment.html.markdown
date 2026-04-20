---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_placement_group_assignment"
sidebar_current: "docs-zenlayercloud-resource-zec_placement_group_assignment"
description: |-
  Provides a resource to assign an instance to a ZEC placement group.
---

# zenlayercloud_zec_placement_group_assignment

Provides a resource to assign an instance to a ZEC placement group.

~> **NOTE:** The number of instances in a placement group must not exceed its partition number (`partition_num`). Exceeding this limit will result in an error.

## Example Usage

```hcl
resource "zenlayercloud_zec_placement_group_assignment" "foo" {
  instance_id        = zenlayercloud_zec_instance.instance.id
  placement_group_id = zenlayercloud_zec_placement_group.pg.id
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, String, ForceNew) The instance ID to assign to the placement group.
* `placement_group_id` - (Required, String, ForceNew) The placement group ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Placement group assignment can be imported using the format `instance_id:placement_group_id`, e.g.

```
terraform import zenlayercloud_zec_placement_group_assignment.foo instance-id:placement-group-id
```

