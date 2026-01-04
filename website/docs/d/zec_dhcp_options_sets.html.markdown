---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_dhcp_options_sets"
sidebar_current: "docs-zenlayercloud-datasource-zec_dhcp_options_sets"
description: |-
  Use this data source to query DHCP Options Sets.
---

# zenlayercloud_zec_dhcp_options_sets

Use this data source to query DHCP Options Sets.

## Example Usage

Query all DHCP Options Sets:

```hcl
data "zenlayercloud_zec_dhcp_options_sets" "all" {

}
```

Query DHCP Options Sets by IDs:

```hcl
data "zenlayercloud_zec_dhcp_options_sets" "by_ids" {
  ids = ["<dphc-options-set-id>"]
}
```

Query DHCP Options Sets by name regex:

```hcl
data "zenlayercloud_zec_dhcp_options_sets" "by_name" {
  name_regex = "^test-.*$"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) List of DHCP options set IDs.
* `name_regex` - (Optional, String) Regular expression for DHCP options set names.
* `resource_group_id` - (Optional, String) ID of the resource group to filter DHCP options sets.
* `result_output_file` - (Optional, String) Output file path.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `dhcp_options_sets` - List of DHCP options sets.
   * `create_time` - Creation time of the DHCP options set.
   * `description` - Description of the DHCP options set.
   * `domain_name_servers` - IPv4 DNS server IP.
   * `id` - DHCP options set ID.
   * `ipv6_domain_name_servers` - IPv6 DNS server IP.
   * `ipv6_lease_time` - IPv6 lease time.
   * `lease_time` - IPv4 lease time.
   * `name` - Name of the DHCP options set.
   * `resource_group_id` - ID of the resource group to which the DHCP options set belongs.
   * `tags` - Tags of the DHCP options set.


