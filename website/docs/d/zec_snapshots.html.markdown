---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_snapshots"
sidebar_current: "docs-zenlayercloud-datasource-zec_snapshots"
description: |-
  Use this data source to query Snapshots
---

# zenlayercloud_zec_snapshots

Use this data source to query Snapshots

## Example Usage

Query all snapshots

```hcl
data "zenlayercloud_zec_snapshots" "all" {}
```

Query snapshots by id

```hcl
# Create a snapshot
resource "zenlayercloud_zec_snapshot" "snapshot" {
  disk_id = "<diskId>"
  name    = "example-snapshot"
}

# Query snapshots using data source
data "zenlayercloud_zec_snapshots" "foo" {
  ids = [zenlayercloud_zec_snapshot.snapshot.id]
}
```

Query snapshots by name regex

```hcl
data "zenlayercloud_zec_snapshots" "foo" {
  name_regex = "^example"
}
```

Query snapshots by availability zone

```hcl
data "zenlayercloud_zec_snapshots" "foo" {
  availability_zone = "asia-east-1a"
}
```

Query snapshots by snapshot type

```hcl
data "zenlayercloud_zec_snapshots" "foo" {
  snapshot_type = "Auto"
}
```

Query snapshots by disk id

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

resource "zenlayercloud_zec_snapshot" "snapshot" {
  disk_id = "<diskId>"
  name    = "example-snapshot"
}

data "zenlayercloud_zec_snapshots" "foo" {
  disk_id = zenlayercloud_zec_snapshot.snapshot.disk_id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) The availability zone of the snapshot to be queried.
* `disk_ids` - (Optional, List: [`String`]) IDs of the disk to be queried.
* `ids` - (Optional, Set: [`String`]) IDs of the snapshots to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the snapshot name.
* `resource_group_id` - (Optional, String) The ID of resource group grouped snapshot to be queried.
* `result_output_file` - (Optional, String) Used to save results.
* `snapshot_type` - (Optional, List: [`String`]) The type of the snapshot to be queried. Valid values: `Auto`, `Manual`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `snapshots` - An information list of snapshot. Each element contains the following attributes:
   * `availability_zone` - The availability zone of snapshot.
   * `create_time` - Creation time of the snapshot.
   * `disk_ability` - Whether the snapshot can be used to create a disk.
   * `disk_id` - The ID of disk that the snapshot is created from.
   * `disk_type` - The Type of disk that the snapshot is created from. Valid values: `SYSTEM`, `DATA`.
   * `id` - ID of the snapshot.
   * `name` - Name of the snapshot.
   * `resource_group_id` - The ID of resource group grouped snapshot.
   * `resource_group_name` - The Name of resource group grouped snapshot.
   * `status` - Status of snapshot. Valid values: `CREATING`, `AVAILABLE`, `FAILED`, `ROLLING_BACK`, `DELETING`.


