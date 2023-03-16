---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_instance_types"
sidebar_current: "docs-zenlayercloud-datasource-instance_types"
description: |-
  Use this data source to query instances type.
---

# zenlayercloud_instance_types

Use this data source to query instances type.

## Example Usage

```hcl
data "zenlayercloud_instance_types" "foo" {

}

data "zenlayercloud_instance_types" "sel1c1g" {
  availability_zone = "SEL-A"
  cpu_count         = 1
  memory            = 1
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) The available zone that the instance locates at.
* `cpu_count` - (Optional, Int) The number of CPU cores of the instance.
* `instance_charge_type` - (Optional, String) The charge type of instance. Valid values are `POSTPAID`, `PREPAID`. The default is `POSTPAID`.
* `instance_type` - (Optional, String) The instance type of the instance.
* `memory` - (Optional, Int) Instance memory capacity, unit in GB.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instance_type_quotas` - An information list of zone available vm instance types. Each element contains the following attributes:
  * `availability_zone` - The zone id that the vm instance locates at.
  * `cpu_count` - The number of CPU cores of the instance.
  * `instance_type` - Type of the instance.
  * `internet_charge_type` - Internet charge type of the instance, Valid values are `ByBandwidth`, `ByTrafficPackage`, `ByInstanceBandwidth95` and `ByClusterBandwidth95`.
  * `maximum_bandwidth_out` - The maximum public bandwidth of the instance type.
  * `memory` - Instance memory capacity, unit in GB.


