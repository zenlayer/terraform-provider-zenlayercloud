---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_vpc"
sidebar_current: "docs-zenlayercloud-resource-bmc_vpc"
description: |-
  Provide a resource to create a VPC.
---

# zenlayercloud_bmc_vpc

Provide a resource to create a VPC.

## Example Usage

```hcl
data "zenlayercloud_bmc_vpc_regions" "default_region" {

}

resource "zenlayercloud_bmc_vpc" "foo" {
  region     = data.zenlayercloud_bmc_vpc_regions.default_region.regions.0.id
  cidr_block = "10.0.0.0/26"
}
```

## Argument Reference

The following arguments are supported:

* `cidr_block` - (Required, String, ForceNew) A network address block which should be a subnet of the three internal network segments (10.0.0.0/16, 172.16.0.0/12 and 192.168.0.0/16).
* `region` - (Required, String, ForceNew) The ID of region that the vpc locates at.
* `name` - (Optional, String) The name of the vpc.
* `resource_group_id` - (Optional, String) The resource group id the vpc belongs to, default to ID of Default Resource Group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the vpc.
* `resource_group_name` - The resource group name the vpc belongs to, default to Default Resource Group.
* `vpc_status` - Current status of the vpc.


## Import

Vpc instance can be imported, e.g.

```
$ terraform import zenlayercloud_bmc_vpc.test vpc-id
```

