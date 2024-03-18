---
subcategory: "Cloud Networking(SDN)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_sdn_port"
sidebar_current: "docs-zenlayercloud-resource-sdn_port"
description: |-
  Provides a resource to manage datacenter port.
---

# zenlayercloud_sdn_port

Provides a resource to manage datacenter port.

## Example Usage

```hcl
resource "zenlayercloud_sdn_port" "foo" {
  name                 = "my_name"
  datacenter           = "xxxxx-xxxxx-xxxxx"
  remarks              = "Test"
  port_type            = "1G"
  business_entity_name = "John"
}
```

## Argument Reference

The following arguments are supported:

* `business_entity_name` - (Required, String) Your business entity name. The entity name to be used on the Letter of Authorization (LOA).
* `datacenter` - (Required, String, ForceNew) ID of data center.
* `port_type` - (Required, String, ForceNew) Type of port. eg. 1G/10G/40G.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the port. Default is `false`. If set true, the port will be permanently deleted instead of being moved into the recycle bin.
* `name` - (Optional, String) Port name. Up to 255 characters in length are allowed.
* `remarks` - (Optional, String) Description of port.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `connect_status` - The network connectivity state of port.
* `create_time` - Create time of the port.
* `datacenter_name` - The name of datacenter.
* `expired_time` - Expired time of the port.
* `loa_status` - The LOA state.
* `loa_url` - The LOA URL address.
* `port_charge_type` - The charge type of port. Valid values: `PREPAID`, `POSTPAID`.
* `port_status` - The business status of port.


## Import

Port can be imported, e.g.

```
$ terraform import zenlayercloud_sdn_port.foo xxxxxx
```

