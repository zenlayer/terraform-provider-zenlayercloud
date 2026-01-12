---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_dhcp_options_set"
sidebar_current: "docs-zenlayercloud-resource-zec_dhcp_options_set"
description: |-
  Provide a resource to manage DHCP Options Set.
---

# zenlayercloud_zec_dhcp_options_set

Provide a resource to manage DHCP Options Set.

## Example Usage

```hcl
resource "zenlayercloud_zec_dhcp_options_set" "example" {
  name                     = "example-dhcp-options-set"
  domain_name_servers      = "8.8.8.8,8.8.4.4"
  ipv6_domain_name_servers = "2001:4860:4860::8888"
  lease_time               = 24
  ipv6_lease_time          = 24
  description              = "example dhcp options set"
  tags = {
    "test" = ""
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name_servers` - (Required, String) IPv4 DNS server IP, up to 4 IPv4 addresses, separated by commas.
* `name` - (Required, String) Name of the DHCP options set.
* `description` - (Optional, String) Description of the DHCP options set.
* `ipv6_domain_name_servers` - (Optional, String) IPv6 DNS server IP, up to 4 IPv6 addresses.
* `ipv6_lease_time` - (Optional, Int) IPv6 lease time, measured in hour. Value range: 24h~1176h, 87600h(3650d)~175200h(7300d). default value: 24.
* `lease_time` - (Optional, Int) IPv4 lease time, measured in hour. Value range: 24h~1176h, 87600h(3650d)~175200h(7300d), default value: 24.
* `resource_group_id` - (Optional, String) ID of the resource group to which the DHCP options set belongs.
* `tags` - (Optional, Map) Tags of the DHCP options set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Creation time of the DHCP options set.


## Import

DHCP Options Set instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_dhcp_options_set.example dhcp-options-set-id
```

