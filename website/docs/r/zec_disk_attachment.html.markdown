---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_disk_attachment"
sidebar_current: "docs-zenlayercloud-resource-zec_disk_attachment"
description: |-
  Provides a resource to attached ZEC disk to an instance.
---

# zenlayercloud_zec_disk_attachment

Provides a resource to attached ZEC disk to an instance.

## Example Usage

```hcl
resource "zenlayercloud_zec_disk_attachment" "test" {
  disk_id     = "<diskId>"
  instance_id = "<instanceId>"
}
```

# Import

Disk attachment can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zec_disk_attachment.test disk-id : instance-id
```

## Argument Reference

The following arguments are supported:

* `disk_id` - (Required, String, ForceNew) The ID of the Disk.
* `instance_id` - (Required, String, ForceNew) The ID of a ZEC instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



