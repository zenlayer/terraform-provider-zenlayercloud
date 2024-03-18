---
subcategory: "Cloud Networking(SDN)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_sdn_datacenters"
sidebar_current: "docs-zenlayercloud-datasource-sdn_datacenters"
description: |-
  Use this data source to get all sdn data centers available.
---

# zenlayercloud_sdn_datacenters

Use this data source to get all sdn data centers available.

## Example Usage

```hcl
data "zenlayercloud_sdn_datacenters" "all" {
}

data "zenlayercloud_sdn_datacenters" "sel" {
  name_regex = "SEL*"
}
```

## Argument Reference

The following arguments are supported:

* `name_regex` - (Optional, String) A regex string to apply to the datacenter list returned.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `datacenters` - An information list of availability datacenter. Each element contains the following attributes:
  * `address` - The location of the datacenter.
  * `area_name` - The region name, like `Asia Pacific`.
  * `city_name` - The name of city where the datacenter located, like `Singapore`.
  * `country_name` - The name of country, like `Singapore`.
  * `id` - ID of the datacenter, which is a uuid format.
  * `name` - The name of the datacenter, like `AP-Singapore1`, usually not used in api parameter.


