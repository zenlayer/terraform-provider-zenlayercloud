---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_disk"
sidebar_current: "docs-zenlayercloud-resource-disk"
description: |-
  Provide a resource to create data disk.
---

# zenlayercloud_disk

Provide a resource to create data disk.

## Example Usage

```hcl
resource "zenlayercloud_disk" "foo" {
  availability_zone = "SEL-A"
  name              = "SEL-20G"
  disk_size         = 20
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The ID of zone that the disk locates at.
* `disk_size` - (Required, Int, ForceNew) The size of disk. Unit: GB. The minimum value is 20 GB.
* `charge_prepaid_period` - (Optional, Int, ForceNew) The tenancy (time unit is month) of the prepaid disk.
* `charge_type` - (Optional, String, ForceNew) Charge type of disk.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the data disk. Default is `false`. If set true, the disk will be permanently deleted instead of being moved into the recycle bin.
* `name` - (Optional, String) The name of the disk.
* `resource_group_id` - (Optional, String) The resource group id the disk belongs to.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the disk.
* `expired_time` - Expire time of the disk.
* `instance_id` - The ID of instance which the disk attached to.


## Import

Disk instance can be imported, e.g.

```
$ terraform import zenlayercloud_disk.test disk-id
```

