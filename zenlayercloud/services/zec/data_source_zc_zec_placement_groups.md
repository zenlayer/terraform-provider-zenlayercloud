Use this data source to query ZEC placement groups.

Example Usage

Query all placement groups
```hcl
data "zenlayercloud_zec_placement_groups" "foo" {
}
```

Query placement groups by ids
```hcl
data "zenlayercloud_zec_placement_groups" "foo" {
  ids = ["<placementGroupId>"]
}
```

Query placement groups by zone
```hcl
data "zenlayercloud_zec_placement_groups" "foo" {
  zone_id = "asia-east-1a"
}
```

Query placement groups by name regex
```hcl
data "zenlayercloud_zec_placement_groups" "foo" {
  name_regex = "example*"
}
```
