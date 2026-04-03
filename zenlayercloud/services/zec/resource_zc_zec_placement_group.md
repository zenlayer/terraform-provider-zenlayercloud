Provides a ZEC placement group resource.

Example Usage

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

Import

Placement group can be imported using the id, e.g.

```
terraform import zenlayercloud_zec_placement_group.foo placement-group-id
```
