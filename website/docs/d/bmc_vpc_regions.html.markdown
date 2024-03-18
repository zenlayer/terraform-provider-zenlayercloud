---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_vpc_regions"
sidebar_current: "docs-zenlayercloud-datasource-bmc_vpc_regions"
description: |-
  Use this data source to get the available regions for vpc.
---

# zenlayercloud_bmc_vpc_regions

Use this data source to get the available regions for vpc.

## Example Usage

```hcl
data "zenlayercloud_bmc_vpc_regions" "sel-region" {
  availability_zone = "SEL-A"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) The zone that the vpc region contains.
* `region` - (Optional, String) The region that the vpc locates at.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `regions` - An information list of vpc regions. Each element contains the following attributes:
   * `availability_zones` - The zones that the vpc region contains.
   * `id` - The ID of the region.
   * `name` - The name of the region.


