---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_ddos_ip_association"
sidebar_current: "docs-zenlayercloud-resource-bmc_ddos_ip_association"
description: |-
  Provides an DDoS IP resource associated with BMC instance.
---

# zenlayercloud_bmc_ddos_ip_association

Provides an DDoS IP resource associated with BMC instance.

## Example Usage

```hcl
resource "zenlayercloud_bmc_ddos_ip_association" "foo" {
  ddos_ip_id  = "ddosIpIdxxxxxx"
  instance_id = "instanceIdxxxxxx"
}
```

## Argument Reference

The following arguments are supported:

* `ddos_ip_id` - (Required, String, ForceNew) The ID of DDoS IP.
* `instance_id` - (Required, String, ForceNew) The instance id going to bind with the DDoS IP.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



## Import

DDoS IP association can be imported using the id, e.g.

```
$ terraform import zenlayercloud_bmc_ddos_ip_association.bar ddosIpIdxxxxxx:instanceIdxxxxxxx
```

