---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_ddos_ip"
sidebar_current: "docs-zenlayercloud-resource-bmc_ddos_ip"
description: |-
  Provides an DDoS IP resource.
---

# zenlayercloud_bmc_ddos_ip

Provides an DDoS IP resource.

## Example Usage

```hcl
variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_bmc_ddos_ip" "foo" {
  availability_zone = var.availability_zone
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The ID of zone that the DDoS IP locates at.
* `charge_prepaid_period` - (Optional, Int, ForceNew) The tenancy (time unit is month) of the prepaid DDoS IP, NOTE: it only works when DDoS charge_type is set to `PREPAID`.
* `charge_type` - (Optional, String, ForceNew) The charge type of DDoS IP. Valid values are `PREPAID`, `POSTPAID`. The default is `POSTPAID`. Note: `PREPAID` DDoS IP may not allow to delete before expired.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the DDoS IP. Default is `false`. If set true, the DDoS IP will be permanently deleted instead of being moved into the recycle bin.
* `resource_group_id` - (Optional, String) The resource group id the DDoS IP belongs to, default to Default Resource Group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the DDoS IP.
* `expired_time` - Expired time of the DDoS IP.
* `ip_status` - Current status of the DDoS IP.
* `public_ip` - The DDoS IP address.
* `resource_group_name` - The resource group name the DDoS IP belongs to, default to Default Resource Group.


## Import

EIP can be imported using the id, e.g.

```
$ terraform import zenlayercloud_bmc_ddos_ip.foo 123123xxxx
```

