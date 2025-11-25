---
subcategory: "Zenlayer Private DNS(zdns)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zdns_zone"
sidebar_current: "docs-zenlayercloud-resource-zdns_zone"
description: |-
  Use this resource to create a DNS Private zone
---

# zenlayercloud_zdns_zone

Use this resource to create a DNS Private zone

## Example Usage

Create a DNS Private zone

```hcl
resource "zenlayercloud_zdns_zone" "foo" {
  zone_name     = "example.com"
  remark        = "test"
  proxy_pattern = "RECURSION"
}
```

## Argument Reference

The following arguments are supported:

* `zone_name` - (Required, String, ForceNew) The name of the private zone.
* `proxy_pattern` - (Optional, String) The recursive DNS proxy. Valid values: `Zone`: Recursive DNS proxy is disabled. `RECURSION`: Recursive DNS proxy is enabled. Default: `ZONE`.
* `remark` - (Optional, String) Remarks.
* `resource_group_id` - (Optional, String) The resource group id the private zone belongs to, default to Default Resource Group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the private zone.
* `resource_group_name` - The resource group name the private zone belongs to, default to Default Resource Group.


## Import

DNS private zone can be imported, e.g.

```
$ terraform import zenlayercloud_zdns_zone.foo zone-id
```

