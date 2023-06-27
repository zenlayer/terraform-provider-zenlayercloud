---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_image"
sidebar_current: "docs-zenlayercloud-resource-image"
description: |-
  Provides a resource to manage image.
---

# zenlayercloud_image

Provides a resource to manage image.

~> **NOTE:** You have to keep the instance power off if the image is created from instance.

## Example Usage

```hcl
resource "zenlayercloud_image" "foo" {
  image_name        = "web-image-centos"
  instance_id       = "xxxxxx"
  image_description = "create a image by the web server"
}
```

## Argument Reference

The following arguments are supported:

* `image_name` - (Required, String) Image name. Cannot be modified unless recreated.
* `instance_id` - (Required, String, ForceNew) VM instance ID.
* `image_description` - (Optional, String) Image description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `image_size` - Image size.


## Import

Image can be imported, e.g.

```
$ terraform import zenlayercloud_image.foo img-xxxxxxx
```

