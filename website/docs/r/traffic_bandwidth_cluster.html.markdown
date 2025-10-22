---
subcategory: "Traffic"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_traffic_bandwidth_cluster"
sidebar_current: "docs-zenlayercloud-resource-traffic_bandwidth_cluster"
description: |-
  Provides a resource to create bandwidth cluster.
---

# zenlayercloud_traffic_bandwidth_cluster

Provides a resource to create bandwidth cluster.

## Example Usage

Create a BGP bandwidth cluster at Amsterdam, billed by monthly 95th percentile, with 100Mbps commitment bandwidth.

```hcl
resource "zenlayercloud_traffic_bandwidth_cluster" "foo" {
  area_code             = "AMS"
  name                  = "example-bandwidth-cluster"
  network_type          = "BGP"
  internet_charge_type  = "MonthlyPercent95Bandwidth"
  commit_bandwidth_mbps = 100
}
```

## Argument Reference

The following arguments are supported:

* `area_code` - (Required, String, ForceNew) The code of area where the bandwidth located.
* `internet_charge_type` - (Required, String, ForceNew) Network billing method. valid values: `MonthlyPercent95Bandwidth`(for Monthly Burstable 95th billing method), `DayPeakBandwidth`(for Daily Peak billing method).
* `name` - (Required, String) The name of the bandwidth cluster.
* `network_type` - (Required, String, ForceNew) IP network type. The value is required when the billing area for bandwidth cluster is by city. valid values: `BGP`(for BGP network), `Cogent`(for Cogent network),`CN2`(for China Telecom Next Carrier Network), `CMI`(for China Mobile network), `CUG`(for China Unicom network), `CTG`(for China Telecom network).
* `commit_bandwidth_mbps` - (Optional, Int) Bandwidth commitment. Measured in Mbps. Default value: `0`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Creation time of the bandwidth cluster.


## Import

Bandwidth cluster can be imported using the id, e.g.

```
terraform import zenlayercloud_traffic_bandwidth_cluster.foo bandwidth-cluster-id
```

