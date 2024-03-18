---
subcategory: "Zenlayer Global Accelerator(ZGA)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zga_accelerate_regions"
sidebar_current: "docs-zenlayercloud-datasource-zga_accelerate_regions"
description: |-
  Use this data source to get all zga available accelerate regions by origin_region_id.
---

# zenlayercloud_zga_accelerate_regions

Use this data source to get all zga available accelerate regions by origin_region_id.

## Example Usage

```hcl
data "zenlayercloud_zga_origin_regions" "DE" {
  name_regex = "DE"
}

data "zenlayercloud_zga_accelerate_regions" "all" {
  origin_region_id = data.zenlayercloud_zga_origin_regions.DE.regions.0.id
}

data "zenlayercloud_zga_accelerate_regions" "FR" {
  origin_region_id = "FR"
  name_regex       = "US*"
}
```

## Argument Reference

The following arguments are supported:

* `origin_region_id` - (Required, String) ID of the origin region, such as `FR`.
* `name_regex` - (Optional, String) A regex string to apply to the accelerate region list returned.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `regions` - An information list of availability accelerate region. Each element contains the following attributes:
   * `description` - The name of the region, like `Frankfurt`, usually not used in api parameter.
   * `id` - ID of the region, such as `FR`.


