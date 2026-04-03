Provides a resource to assign an instance to a ZEC placement group.

~> **NOTE:** The number of instances in a placement group must not exceed its partition number (`partition_num`). Exceeding this limit will result in an error.

Example Usage

```hcl

resource "zenlayercloud_zec_placement_group_assignment" "foo" {
  instance_id        = zenlayercloud_zec_instance.instance.id
  placement_group_id = zenlayercloud_zec_placement_group.pg.id
}

```

Import

Placement group assignment can be imported using the format `instance_id:placement_group_id`, e.g.

```
terraform import zenlayercloud_zec_placement_group_assignment.foo instance-id:placement-group-id
```
