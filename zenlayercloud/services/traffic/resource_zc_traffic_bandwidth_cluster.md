Provides a resource to create bandwidth cluster.

Example Usage

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

Import

Bandwidth cluster can be imported using the id, e.g.

```
terraform import zenlayercloud_traffic_bandwidth_cluster.foo bandwidth-cluster-id
```
