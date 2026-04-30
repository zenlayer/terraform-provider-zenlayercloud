---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_image"
sidebar_current: "docs-zenlayercloud-resource-zec_image"
description: |-
  Provides a resource to manage a custom image in Zenlayer Elastic Compute (ZEC).
---

# zenlayercloud_zec_image

Provides a resource to manage a custom image in Zenlayer Elastic Compute (ZEC).

Creating this resource will make a custom image from an existing ZEC instance.
Updating `image_name` calls `ModifyImagesAttributes` in place.

## Example Usage

Create a custom image from an instance

```hcl
resource "zenlayercloud_zec_image" "foo" {
  instance_id = "1660545330971680835"
  image_name  = "my-custom-image"

  tags = {
    env = "test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `image_name` - (Required, String) Name of the image. 2-63 chars; letters, digits, `-`, `_`, `.`; must start and end with a letter or digit.
* `instance_id` - (Optional, String, ForceNew) ID of the ZEC instance to create the image from. Required for creation; ignored on import.
* `resource_group_id` - (Optional, String) Resource group the image belongs to. Defaults to the default resource group.
* `tags` - (Optional, Map) Tags bound to the image.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `category` - Image catalog, e.g. `CentOS`, `Ubuntu`.
* `image_description` - Description of the image.
* `image_size` - Size of the image, in GiB.
* `image_source` - Source of the image.
* `image_status` - Status of the image.
* `image_type` - Type of the image. Typically `CUSTOM_IMAGE` for resources created here.
* `image_version` - OS version of the image.
* `nic_network_type` - Supported NIC network types.
* `os_type` - OS type of the image, such as `windows` or `linux`.


## Import

A custom image can be imported using its id, e.g.

```
$ terraform import zenlayercloud_zec_image.foo <imageId>
```

