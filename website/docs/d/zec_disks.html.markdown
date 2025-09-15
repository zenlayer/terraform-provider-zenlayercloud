---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_disks"
sidebar_current: "docs-zenlayercloud-datasource-zec_disks"
description: |-
  Use this data source to query zec disk information.
---

# zenlayercloud_zec_disks

Use this data source to query zec disk information.

## Example Usage

Query all disks storages

```hcl
data "zenlayercloud_zec_disks" "all" {
}
```

Query disks by availability zone

```hcl
data "zenlayercloud_zec_disks" "zone_disk" {
  availability_zone = "asia-east-1"
}
```

Query disks by ids

```hcl
data "zenlayercloud_zec_disks" "zone_disk" {
  ids = ["<diskId>"]
}
```

Query disks by disk type

```hcl
data "zenlayercloud_zec_disks" "system_disk" {
  disk_type = "SYSTEM"
}
```

Query disks by attached instance

```hcl
data "zenlayercloud_zec_disks" "instance_disk" {
  instance_id = "<instanceId>"
}
```

Query disks by name regex

```hcl
data "zenlayercloud_zec_disks" "name_disk" {
  name_regex = "disk20*"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) Zone of the disk to be queried.
* `disk_type` - (Optional, String) Type of the disk. Valid values: `SYSTEM`, `DATA`.
* `ids` - (Optional, Set: [`String`]) ids of the disk to be queried.
* `instance_id` - (Optional, String) Query the disks which attached to the instance.
* `name_regex` - (Optional, String) A regex string to apply to the disk list returned.
* `resource_group_id` - (Optional, String) The ID of resource group grouped disk to be queried.
* `result_output_file` - (Optional, String) Used to save results.
* `status` - (Optional, String) Status of disk to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `disks` - An information list of disk. Each element contains the following attributes:
   * `auto_snapshot_policy_id` - The ID of auto snapshot policy associated with this disk.
   * `availability_zone` - The availability zone of disk.
   * `create_time` - Creation time of the disk.
   * `disk_category` - The category of disk.
   * `disk_size` - Size of the disk.
   * `disk_type` - Type of the disk. Values are: `SYSTEM`, `DATA`.
   * `id` - ID of the disk.
   * `instance_id` - The ID of instance that the disk attached to.
   * `instance_name` - The name of instance that the disk attached to.
   * `name` - name of the disk.
   * `resource_group_id` - The ID of resource group grouped disk to be queried.
   * `resource_group_name` - The Name of resource group grouped disk to be queried.
   * `status` - Status of disk.


