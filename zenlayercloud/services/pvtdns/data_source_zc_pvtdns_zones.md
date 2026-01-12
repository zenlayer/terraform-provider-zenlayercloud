Use this data source to query DNS private zones

Example Usage

Query all DNS private zones

```hcl
data "zenlayercloud_pvtdns_zones" "all" {
}
```

Query DNS private zones by ids

```hcl
data "zenlayercloud_pvtdns_zones" "foo" {
  ids = ["<zoneId>"]
}
```

Query DNS private zones by zone name regex

```hcl
data "zenlayercloud_pvtdns_zones" "foo" {
  name_regex = "test*"
}
```

Query DNS private zones by resource group id 

```hcl
data "zenlayercloud_pvtdns_zones" "foo" {
  resource_group_id = "xxxx"
}
```
