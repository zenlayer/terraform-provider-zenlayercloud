Use this data source to query the bandwidth cluster areas

Example Usage

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
