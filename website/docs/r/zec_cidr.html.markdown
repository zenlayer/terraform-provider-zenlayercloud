---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_cidr"
sidebar_current: "docs-zenlayercloud-resource-zec_cidr"
description: |-
  Provide a resource to create CIDR block.
---

# zenlayercloud_zec_cidr

Provide a resource to create CIDR block.

## Example Usage

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_cidr" "test" {
  region_id    = var.region
  netmask      = 27
  network_type = "BGPLine"
}
```

# Import

CIDR block can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zec_cidr.test cidr-block-id
```

## Argument Reference

The following arguments are supported:

* `netmask` - (Required, Int, ForceNew) Netmask of CIDR block. Valid values: `27` to `30`.
* `region_id` - (Required, String, ForceNew) The region ID that the public CIDR block locates at.
* `name` - (Optional, String) Name of the public CIDR block.The name should start and end with a number or a letter, containing 2 to 63 characters. Only letters, numbers, -, slash(/) and periods (.) are supported.
* `network_type` - (Optional, String, ForceNew) Network types of public CIDR block. Valid values: `BGPLine`, `CN2Line`, `LocalLine`, `ChinaTelecom`, `ChinaUnicom`, `ChinaMobile`, `Cogent`.
* `resource_group_id` - (Optional, String) Resource group ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `cidr_block_address` - The public CIDR block address.
* `create_time` - Creation time of the public CIDR block.
* `resource_group_name` - The Name of resource group.
* `status` - Status of the public CIDR block.


