---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_disk_snapshot"
sidebar_current: "docs-zenlayercloud-resource-zec_disk_snapshot"
description: |-
  Provides a resource to create snapshot for ZEC disk.
---

# zenlayercloud_zec_disk_snapshot

Provides a resource to create snapshot for ZEC disk.

## Example Usage

Prepare a disk

```hcl
variable "availability_zone" {
  default = "asia-east-1a"
}

resource "zenlayercloud_zec_disk" "test" {
  availability_zone = var.availability_zone
  disk_name         = "Disk-20G"
  disk_size         = 60
  disk_category     = "Standard NVMe SSD"
}
```

Create a snapshot

```hcl
resource "zenlayercloud_zec_disk_snapshot" "snapshot" {
  disk_id = zenlayercloud_zec_disk.test.id
  name    = "example-snapshot"
}
```

## Argument Reference

The following arguments are supported:

* `disk_id` - (Required, String, ForceNew) The ID of disk which the snapshot created from.
* `name` - (Optional, String) The name of the snapshot. The name should start and end with a number or a letter, containing 2 to 63 characters. Only letters, numbers, - and periods (.) are supported.
* `retention_time` - (Optional, String) Retention time of snapshot. Valid format: yyyy-MM-ddTHH:mm:ssZ, and must be at least 24 hours in the future. Example: 2025-10-01T10:10:10Z.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `availability_zone` - The availability zone of snapshot.
* `create_time` - Creation time of the snapshot.
* `disk_ability` - Whether the snapshot can be used to create a disk.
* `resource_group_id` - The ID of resource group grouped snapshot.
* `resource_group_name` - The Name of resource group grouped snapshot.
* `snapshot_type` - The type of the snapshot to be queried. Valid values: `Auto`, `Manual`.
* `status` - Status of snapshot. Valid values: `CREATING`, `AVAILABLE`, `FAILED`, `ROLLING_BACK`, `DELETING`.


## Import

Snapshot can be imported, e.g.

```
$ terraform import zenlayercloud_zec_disk_snapshot.snapshot snapshot-id
```

