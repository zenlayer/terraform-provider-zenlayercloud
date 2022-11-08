---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_eip_association"
sidebar_current: "docs-zenlayercloud-resource-bmc_eip_association"
description: |-
  Provides an eip resource associated with BMC instance.
---

# zenlayercloud_bmc_eip_association

Provides an eip resource associated with BMC instance.

## Example Usage

```hcl
resource "zenlayercloud_bmc_eip_association" "foo" {
  eip_id      = "eipxxxxxx"
  instance_id = "instanceIdxxxxxx"
}
```

## Argument Reference

The following arguments are supported:

* `eip_id` - (Required, String, ForceNew) The ID of EIP.
* `instance_id` - (Required, String, ForceNew) The instance id going to bind with the EIP.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

Eip association can be imported using the id, e.g.

```
$ terraform import zenlayercloud_bmc_eip_association.bar eipIdxxxxxx:instanceIdxxxxxxx
```

