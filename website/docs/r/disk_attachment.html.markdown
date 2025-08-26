---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_disk_attachment"
sidebar_current: "docs-zenlayercloud-resource-disk_attachment"
description: |-
  Provide a resource to attach a disk to an instance.
---

# zenlayercloud_disk_attachment

Provide a resource to attach a disk to an instance.

## Example Usage

```hcl
resource "zenlayercloud_disk_attachment" "foo" {
  disk_id     = "diskxxxx"
  instance_id = "instancexxxx"
}
```

## Argument Reference

The following arguments are supported:

* `disk_id` - (Required, String, ForceNew) The ID of disk.
* `instance_id` - (Required, String, ForceNew) The ID of instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Disk attachment can be imported, e.g.

```
$ terraform import zenlayercloud_disk_attachment.foo disk-id:instance-id
```

