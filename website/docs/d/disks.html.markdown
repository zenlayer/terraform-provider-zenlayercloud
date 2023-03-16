---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_disks"
sidebar_current: "docs-zenlayercloud-datasource-disks"
description: |-
  Use this data source to query vm disk information.
---

# zenlayercloud_disks

Use this data source to query vm disk information.

## Example Usage

```hcl
data "zenlayercloud_disks" "all" {
}

# filter system disk
data "zenlayercloud_disks" "system_disk" {
  disk_type = "SYSTEM"
}

#filter with name regex
data "zenlayercloud_disks" "name_disk" {
  name_regex = "disk20*"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) Zone of the disk to be queried.
* `disk_type` - (Optional, String) Type of the disk. Valid values: `SYSTEM`, `DATA`.
* `id` - (Optional, String) id of the disk to be queried.
* `instance_id` - (Optional, String) Query the disks which attached to the instance.
* `name_regex` - (Optional, String) A regex string to apply to the disk list returned.
* `name` - (Optional, String) Fuzzy query with this name.
* `portable` - (Optional, Bool) Whether the disk is deleted with instance or not, true means not delete with instance, false otherwise.
* `result_output_file` - (Optional, String) Used to save results.
* `status` - (Optional, String) Status of disk to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `disks` - An information list of disk. Each element contains the following attributes:
  * `availability_zone` - The availability zone of disk.
  * `charge_type` - Charge type of the disk. Values are: `PREPAID`, `POSTPAID`.
  * `create_time` - Creation time of the disk.
  * `disk_category` - The category of disk. Values are: cloud_efficiency.
  * `disk_size` - Size of the disk.
  * `disk_type` - Type of the disk. Values are: `SYSTEM`, `DATA`.
  * `expired_time` - Expired Time of the disk.
  * `id` - ID of the disk.
  * `instance_id` - The ID of instance that the disk attached to.
  * `instance_name` - The name of instance that the disk attached to.
  * `name` - name of the disk.
  * `period` - The period cycle of the disk. Unit: month.
  * `portable` - Whether the disk is deleted with instance or not, true means not delete with instance, false otherwise.
  * `status` - Status of disk.


