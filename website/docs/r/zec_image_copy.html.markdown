---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_image_copy"
sidebar_current: "docs-zenlayercloud-resource-zec_image_copy"
description: |-
  Provides a resource to manage the cross-region distribution of a custom ZEC image.
---

# zenlayercloud_zec_image_copy

Provides a resource to manage the cross-region distribution of a custom ZEC image.

Declare the full set of regions (including the source region) where the image
should exist; Terraform will call CopyImage for additions and DeleteImageCopy
for removals.

## Example Usage

Distribute a custom image to multiple regions

```hcl
resource "zenlayercloud_zec_image_copy" "example" {
  image_id       = zenlayercloud_zec_image.img.id
  region_id_list = ["SHA", "SEL", "FRA"]
}
```

## Argument Reference

The following arguments are supported:

* `image_id` - (Required, String, ForceNew) ID of the custom image to distribute.
* `region_id_list` - (Required, Set: [`String`]) Full set of region IDs where the image should exist (including the source region). At least one region must remain at all times.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

An image-copy resource can be imported using the image id, e.g.

```
$ terraform import zenlayercloud_zec_image_copy.example <imageId>
```

