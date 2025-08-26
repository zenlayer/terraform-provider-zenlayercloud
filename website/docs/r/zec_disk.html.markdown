---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_disk"
sidebar_current: "docs-zenlayercloud-resource-zec_disk"
description: |-
  Provide a resource to create data disk.
---

# zenlayercloud_zec_disk

Provide a resource to create data disk.

## Example Usage

```hcl
variable "availability_zone" {
  default = "asia-east-1a"
}

resource "zenlayercloud_zec_disk" "test" {
  availability_zone = "SH1A"
  disk_name         = "Disk-20G"
  disk_size         = 60
  disk_category     = "Standard NVMe SSD"
}
```

# Import

Disk instance can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zec_disk.test disk-id
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The ID of zone that the disk locates at.
* `disk_size` - (Required, Int) The size of disk. Unit: GB. The minimum value is 20 GB. When resize the disk, the new size must be greater than the former value.
* `disk_category` - (Optional, String, ForceNew) The category of disk.
* `disk_name` - (Optional, String) The name of the disk.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the data disk. Default is `false`. If set true, the disk will be permanently deleted instead of being moved into the recycle bin.
* `resource_group_id` - (Optional, String) The resource group id the disk belongs to.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the disk.
* `disk_type` - Type of the disk. Values are: `SYSTEM`, `DATA`.


