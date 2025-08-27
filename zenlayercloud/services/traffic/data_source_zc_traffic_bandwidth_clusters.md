Use this data source to query the bandwidth cluster instances.

Example Usage

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
