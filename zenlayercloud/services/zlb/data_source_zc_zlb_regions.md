Use this data source to query available regions for load balancer.

Example Usage

Query all load balancer regions
```hcl
data "zenlayercloud_zlb_regions" "all" {
}
```

Query load balancer regions by city code
```hcl
data "zenlayercloud_zlb_regions" "foo" {
	city_code = "SEL"s
}
```
