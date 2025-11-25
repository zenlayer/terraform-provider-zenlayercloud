---
subcategory: "Zenlayer Private DNS(zdns)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zdns_zones"
sidebar_current: "docs-zenlayercloud-datasource-zdns_zones"
description: |-
  Use this data source to query DNS private zones
---

# zenlayercloud_zdns_zones

Use this data source to query DNS private zones

## Example Usage

Query all DNS private zones

```hcl
data "zenlayercloud_zdns_zones" "all" {
}
```

Query DNS private zones by ids

```hcl
data "zenlayercloud_zdns_zones" "foo" {
  ids = ["<zoneId>"]
}
```

Query DNS private zones by zone name regex

```hcl
data "zenlayercloud_zdns_zones" "foo" {
  name_regex = "test*"
}
```

Query DNS private zones by resource group id

```hcl
data "zenlayercloud_zdns_zones" "foo" {
  resource_group_id = "xxxx"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) IDs of the private DNS zones to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the private DNS zone list returned.
* `resource_group_id` - (Optional, String) Resource group ID.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `zones` - An information list of private DNS zones. Each element contains the following attributes:
   * `create_time` - Creation time of the private DNS zone.
   * `id` - ID of the private DNS zone.
   * `proxy_pattern` - Indicate whether the recursive resolution proxy is enabled or disabled.
   * `remark` - Remark of the private DNS zone.
   * `resource_group_id` - ID of resource group.
   * `resource_group_name` - Resource group name.
   * `tags` - tags.
      * `tag_key` - tag key.
      * `tag_value` - tag value.
   * `zone_name` - Name of the private DNS zone.


