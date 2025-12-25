---
subcategory: "Zenlayer Private DNS(ZDNS)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zdns_zone"
sidebar_current: "docs-zenlayercloud-resource-zdns_zone"
description: |-
  Use this resource to create a DNS Private zone
---

# zenlayercloud_zdns_zone

Use this resource to create a DNS Private zone

For more information about Zenlayer DNS, see the Zenlayer Documentation on [ZDNS Service](https://docs.console.zenlayer.com/welcome/elastic-compute/overview/zdns-service)

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
* `proxy_pattern` - (Optional, String) The recursive DNS proxy setting for subdomains. Default: `ZONE`. Valid values: 
	- `ZONE`: Disable recursive DNS proxy. When resolving non-existent subdomains under this domain, it directly returns NXDOMAIN, indicating the subdomain does not exist. 
	- `RECURSION`: Enable recursive DNS proxy. When resolving non-existent subdomains under this domain, it queries the recursive module and responds to the resolution request with the final query result.
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

