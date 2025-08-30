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
resource "zenlayercloud_zec_disk_snapshot_policy_attachment" "test" {
  disk_id                 = "<diskId>"
  auto_snapshot_policy_id = "<autoSnapshotPolicyId>"
}
```

# Import

Disk Snapshot Policy attachment can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zec_disk_snapshot_policy_attachment.test disk-id
```

## Argument Reference

The following arguments are supported:

* `auto_snapshot_policy_id` - (Required, String, ForceNew) The ID of the auto snapshot policy.
* `disk_id` - (Required, String, ForceNew) The ID of the disk.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



