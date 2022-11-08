---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_eip"
sidebar_current: "docs-zenlayercloud-resource-bmc_eip"
description: |-
  Provides an EIP resource.
---

# zenlayercloud_bmc_eip

Provides an EIP resource.

## Example Usage

```hcl
variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_bmc_eip" "foo" {
  availability_zone = var.availability_zone
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The ID of zone that the EIP locates at.
* `eip_charge_prepaid_period` - (Optional, Int, ForceNew) The tenancy (time unit is month) of the prepaid EIP, NOTE: it only works when eip_charge_type is set to `PREPAID`.
* `eip_charge_type` - (Optional, String, ForceNew) The charge type of EIP. Valid values are `PREPAID`, `POSTPAID`. The default is `POSTPAID`. Note: `PREPAID` EIP may not allow to delete before expired.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the EIP. Default is `false`. If set true, the EIP will be permanently deleted instead of being moved into the recycle bin.
* `resource_group_id` - (Optional, String) The resource group id the EIP belongs to, default to Default Resource Group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the EIP.
* `eip_status` - Current status of the EIP.
* `expired_time` - Expired time of the EIP.
* `public_ip` - The EIP address.
* `resource_group_name` - The resource group name the EIP belongs to, default to Default Resource Group.


## Import

EIP can be imported using the id, e.g.

```
$ terraform import zenlayercloud_bmc_eip.foo 123123xxxx
```

