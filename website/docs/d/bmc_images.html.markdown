---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_images"
sidebar_current: "docs-zenlayercloud-datasource-bmc_images"
description: |-
  Use this data source to query images.
---

# zenlayercloud_bmc_images

Use this data source to query images.

## Example Usage

```hcl
data "zenlayercloud_bmc_images" "foo" {
  catalog          = "centos"
  instance_type_id = "S9I"
}
```

## Argument Reference

The following arguments are supported:

* `catalog` - (Optional, String) The catalog which the image belongs to. Valid values: 'centos', 'windows', 'ubuntu', 'debian', 'esxi'.
* `image_id` - (Optional, String) ID of the image.
* `image_name` - (Optional, String) Name of the image, such as `CentOS7.4-x86_64`.
* `image_type` - (Optional, String) The image type. Valid values: 'PUBLIC_IMAGE', 'CUSTOM_IMAGE'.
* `instance_type_id` - (Optional, String) Filter images which are supported to install on specified instance type, such as `M6C`.
* `os_type` - (Optional, String) os type of the image. Valid values: 'windows', 'linux'.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `images` - An information list of image. Each element contains the following attributes:
   * `catalog` - Created time of the image.
   * `image_id` - ID of the image.
   * `image_name` - Name of the image.
   * `image_type` - Type of the image. with value: `PUBLIC_IMAGE` and `CUSTOM_IMAGE`.
   * `os_type` - Type of the image, windows or linux.


