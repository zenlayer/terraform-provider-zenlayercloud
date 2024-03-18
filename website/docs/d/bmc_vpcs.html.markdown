---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_vpcs"
sidebar_current: "docs-zenlayercloud-datasource-bmc_vpcs"
description: |-
  Use this data source to query vpc information.
---

# zenlayercloud_bmc_vpcs

Use this data source to query vpc information.

## Example Usage

```hcl
data "zenlayercloud_bmc_vpc_regions" "region" {
}

resource "zenlayercloud_bmc_vpc" "foo" {
  region     = data.zenlayercloud_bmc_vpc_regions.region.vpc_regions.0.region
  name       = "test_vpc"
  cidr_block = "10.0.0.0/16"
}

data "zenlayercloud_bmc_vpcs" "foo" {
}
```

## Argument Reference

The following arguments are supported:

* `cidr_block` - (Optional, String) Filter VPC with this CIDR.
* `region` - (Optional, String) region of the VPC to be queried.
* `resource_group_id` - (Optional, String) The ID of resource group grouped VPC to be queried.
* `result_output_file` - (Optional, String) Used to save results.
* `vpc_id` - (Optional, String) ID of the VPC to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `vpc_list` - An information list of VPC. Each element contains the following attributes:
   * `cidr_block` - A network address block of the VPC.
   * `create_time` - Creation time of the VPC.
   * `name` - Name of the VPC.
   * `region` - The region where the VPC located.
   * `resource_group_id` - The ID of resource group grouped VPC to be queried.
   * `resource_group_name` - The Name of resource group grouped VPC to be queried.
   * `vpc_id` - ID of the VPC.
   * `vpc_status` - status of the VPC.


