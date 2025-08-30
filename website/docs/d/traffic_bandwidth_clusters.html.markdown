---
subcategory: "Traffic"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_traffic_bandwidth_clusters"
sidebar_current: "docs-zenlayercloud-datasource-traffic_bandwidth_clusters"
description: |-
  Use this data source to query the bandwidth cluster instances.
---

# zenlayercloud_traffic_bandwidth_clusters

Use this data source to query the bandwidth cluster instances.

## Example Usage

Query all bandwidth cluster areas

```hcl
data "zenlayercloud_traffic_bandwidth_clusters" "all" {
}
```

Filter bandwidth cluster areas by id

```hcl
data "zenlayercloud_traffic_bandwidth_clusters" "foo" {
  ids = ["bandwidthClusterId"]
}
```

Filter bandwidth cluster by city name

```hcl
data "zenlayercloud_traffic_bandwidth_clusters" "foo" {
  city_name = "Shanghai"
}
```

Filter bandwidth cluster by name regex

```hcl
data "zenlayercloud_traffic_bandwidth_clusters" "foo" {
  name_regex = "BGP-Shanghai*"
}
```

## Argument Reference

The following arguments are supported:

* `city_name` - (Optional, String) Name of city where the bandwidth cluster located.
* `ids` - (Optional, Set: [`String`]) ids of the bandwidth cluster to be queried.
* `name_regex` - (Optional, String) A regex string to apply to the bandwidth cluster list returned.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `bandwidth_clusters` - An information list of bandwidth cluster. Each element contains the following attributes:
   * `area_code` - The code of area where the bandwidth located.
   * `commit_bandwidth_mbps` - Bandwidth commitment. Measured in Mbps.
   * `create_time` - Creation time of the bandwidth cluster.
   * `id` - ID of the bandwidth cluster.
   * `internet_charge_type` - Network billing method. valid values: `MonthlyPercent95Bandwidth`(for Monthly Burstable 95th billing method), `DayPeakBandwidth`(for Daily Peak billing method).
   * `name` - The name of the bandwidth cluster.
   * `network_type` - IP network type. valid values: `BGP`(for BGP network), `Cogent`(for Cogent network),`CN2`(for China Telecom Next Carrier Network), `CMI`(for China Mobile network), `CUG`(for China Unicom network), `CTG`(for China Telecom network).


