---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_images"
sidebar_current: "docs-zenlayercloud-datasource-images"
description: |-
  Use this data source to query images.
---

# zenlayercloud_images

Use this data source to query images.

## Example Usage

```hcl
data "zenlayercloud_images" "foo" {
  availability_zone = "FRA-A"
  category          = "CentOS"
  image_type        = ["PUBLIC_IMAGE"]
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String) Zone of the images to be queried.
* `category` - (Optional, String) The catalog which the image belongs to. Valid values: 'CentOS', 'Windows', 'Ubuntu', 'Debian'.
* `image_id` - (Optional, String) ID of the image.
* `image_name_regex` - (Optional, String) A regex string to apply to the image list returned by ZenlayerCloud, conflict with 'os_name'. **NOTE**: it is not wildcard, should look like `image_name_regex = "^CentOS\s+6\.8\s+64\w*"`.
* `image_type` - (Optional, String) The image type. Valid values: 'PUBLIC_IMAGE', 'CUSTOM_IMAGE'.
* `os_type` - (Optional, String) os type of the image. Valid values: 'windows', 'linux'.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `images` - An information list of image. Each element contains the following attributes:
  * `category` - The catalog which the image belongs to. With values: 'CentOS', 'Windows', 'Ubuntu', 'Debian'.
  * `image_description` - The description of image.
  * `image_id` - ID of the image.
  * `image_name` - Name of the image.
  * `image_size` - The size of image.
  * `image_type` - Type of the image. With value: `PUBLIC_IMAGE` and `CUSTOM_IMAGE`.
  * `image_version` - The version of image, such as 2019.
  * `os_type` - Type of the image, windows or linux.


