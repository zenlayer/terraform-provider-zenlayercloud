---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_disk_snapshot_policy_attachment"
sidebar_current: "docs-zenlayercloud-resource-zec_disk_snapshot_policy_attachment"
description: |-
  Provides a resource to attached ZEC disk to an auto snapshot policy.
---

# zenlayercloud_zec_disk_snapshot_policy_attachment

Provides a resource to attached ZEC disk to an auto snapshot policy.

## Example Usage

```hcl
var "availability_zone" {
  default = "asia-east-1a"
}

resource "zenlayercloud_zec_disk" "test" {
  availability_zone = var.availability_zone
  disk_name         = "Disk-20G"
  disk_size         = 60
  disk_category     = "Standard NVMe SSD"
}

resource "zenlayercloud_zec_disk_snapshot_policy" "test" {
  availability_zone = var.availability_zone
  name              = "example-snapshot-policy"
  repeat_week_days  = [1]
  hours             = [12]
  retention_days    = 7
}

resource "zenlayercloud_zec_disk_snapshot_policy_attachment" "test" {
  disk_id                 = zenlayercloud_zec_disk.test.id
  auto_snapshot_policy_id = zenlayercloud_zec_disk_snapshot_policy.test.id
}
```

## Argument Reference

The following arguments are supported:

* `auto_snapshot_policy_id` - (Required, String, ForceNew) The ID of the auto snapshot policy.
* `disk_id` - (Required, String, ForceNew) The ID of the disk. Note: system disk is not support yet.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Disk Snapshot Policy attachment can be imported, e.g.

```
$ terraform import zenlayercloud_zec_disk_snapshot_policy_attachment.test disk-id
```

