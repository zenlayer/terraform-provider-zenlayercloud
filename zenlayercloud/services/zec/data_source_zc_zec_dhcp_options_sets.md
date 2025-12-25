Use this data source to query DHCP Options Sets.

Example Usage

Query all DHCP Options Sets:
```hcl
data "zenlayercloud_zec_dhcp_options_sets" "all"{
  
}
```

Query DHCP Options Sets by IDs:
```hcl
data "zenlayercloud_zec_dhcp_options_sets" "by_ids" {
  ids = ["<dphc-options-set-id>"]
}
```

Query DHCP Options Sets by name regex:
```hcl
data "zenlayercloud_zec_dhcp_options_sets" "by_name" {
  name_regex = "^test-.*$"
}
```
