---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_images"
sidebar_current: "docs-zenlayercloud-datasource-zec_images"
description: |-
  Use this data source to query images.
---

# zenlayercloud_zec_images

Use this data source to query images.

## Example Usage

```hcl
variable "availability_zone" {
  default = "asia-east-1a"
}

data "zenlayercloud_images" "foo" {
  availability_zone = var.availability_zone
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String) Zone of the images to be queried.
* `category` - (Optional, String) The catalog which the image belongs to. such as `CentOS`, `Windows`, `FreeBSD` etc.
* `ids` - (Optional, Set: [`String`]) IDs of the image to be queried.
* `image_name_regex` - (Optional, String) A regex string to apply to the image list returned by ZenlayerCloud, conflict with 'os_name'. **NOTE**: it is not wildcard, should look like `image_name_regex = "^CentOS\s+6\.8\s+64\w*"`.
* `image_type` - (Optional, String) The image type. Valid values: 'PUBLIC_IMAGE', 'CUSTOM_IMAGE'.
* `os_type` - (Optional, String) os type of the image. Valid values: 'windows', 'linux', 'bsd', 'android', 'any'.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `images` - An information list of image. Each element contains the following attributes:
   * `category` - The catalog which the image belongs to. With values: 'CentOS', 'Windows', 'Ubuntu', 'Debian'.
   * `id` - ID of the image.
   * `image_description` - The description of image.
   * `image_name` - Name of the image.
   * `image_size` - The size of image. Measured in GiB.
   * `image_type` - Type of the image. With value: `PUBLIC_IMAGE` and `CUSTOM_IMAGE`.
   * `image_version` - The version of image, such as 'Server 20.04 LTS'.
   * `os_type` - Type of the image, `windows` or `linux`.


