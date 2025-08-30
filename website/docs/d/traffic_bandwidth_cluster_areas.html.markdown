---
subcategory: "Traffic"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_traffic_bandwidth_cluster_areas"
sidebar_current: "docs-zenlayercloud-datasource-traffic_bandwidth_cluster_areas"
description: |-
  Use this data source to query the bandwidth cluster areas
---

# zenlayercloud_traffic_bandwidth_cluster_areas

Use this data source to query the bandwidth cluster areas

## Example Usage

Query all bandwidth cluster areas

```hcl
data "zenlayercloud_traffic_bandwidth_cluster_areas" "all" {
}
```

Filter bandwidth cluster areas by area code

```hcl
data "zenlayercloud_traffic_bandwidth_cluster_areas" "foo" {
  area_code = "SHA"
}
```

Filter bandwidth cluster areas by network type

```hcl
data "zenlayercloud_traffic_bandwidth_cluster_areas" "foo" {
  network_type = "BGP"
}
```

Filter bandwidth cluster areas by name regex

```hcl
data "zenlayercloud_traffic_bandwidth_cluster_areas" "foo" {
  name_regex = "shanghai*"
}
```

## Argument Reference

The following arguments are supported:

* `area_code` - (Optional, String) Code(ID) of the bandwidth cluster area.
* `name_regex` - (Optional, String) A regex string to apply to the name of bandwidth cluster area list returned.
* `network_type` - (Optional, String) The IP network support to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `areas` - An information list of bandwidth cluster areas. Each element contains the following attributes:
   * `area_code` - ID of the bandwidth cluster area.
   * `name` - The name of the bandwidth cluster area.
   * `network_types` - IP network type support in the bandwidth cluster. valid values: `BGP`(for BGP network), `Cogent`(for Cogent network),`CN2`(for China Telecom Next Carrier Network), `CMI`(for China Mobile network), `CUG`(for China Unicom network), `CTG`(for China Telecom network).


