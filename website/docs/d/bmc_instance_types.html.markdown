---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_instance_types"
sidebar_current: "docs-zenlayercloud-datasource-bmc_instance_types"
description: |-
  Use this data source to query instances types.
---

# zenlayercloud_bmc_instance_types

Use this data source to query instances types.

## Example Usage

```hcl
data "zenlayercloud_bmc_instance_types" "foo" {

}

data "zenlayercloud_bmc_instance_types" "sel" {
  availability_zone    = "SEL-A"
  instance_charge_type = "PREPAID"
  exclude_sold_out     = true
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) The available zone that the BMC instance locates at.
* `exclude_sold_out` - (Optional, Bool) Indicate to filter instances types that is sold out or not, default is false.
* `instance_charge_type` - (Optional, String) The charge type of instance. Valid values are `POSTPAID`, `PREPAID`. The default is `POSTPAID`.
* `instance_type_id` - (Optional, String) The instance type id of the instance.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instance_types` - An information list of available bmc instance types. Each element contains the following attributes:
   * `availability_zone` - The zone id that the bmc instance locates at.
   * `default_traffic_package_size` - The default value of traffic package size.
   * `instance_type_id` - Type ID of the instance.
   * `internet_charge_types` - The supported internet charge types of the instance at specified zone.
   * `maximum_bandwidth_out` - The maximum public bandwidth of the instance type.
   * `sell_status` - Sell status of the instance.


