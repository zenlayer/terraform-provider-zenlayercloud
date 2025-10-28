---
subcategory: "Zenlayer Virtual Machine(ZVM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zvm_zones"
sidebar_current: "docs-zenlayercloud-datasource-zvm_zones"
description: |-
  Use this data source to get all vm available zones.
---

# zenlayercloud_zvm_zones

Use this data source to get all vm available zones.

## Example Usage

```hcl
data "zenlayercloud_zvm_zones" "all" {
}

data "zenlayercloud_zvm_zones" "sel" {
  name_regex = "SEL*"
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional, String) A regex string to apply to the zone list returned.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `zones` - An information list of availability zone. Each element contains the following attributes:
   * `description` - The name of the zone, like `Frankfurt`, usually not used in api parameter.
   * `name` - ID of the zone, such as `FRA-A`.


