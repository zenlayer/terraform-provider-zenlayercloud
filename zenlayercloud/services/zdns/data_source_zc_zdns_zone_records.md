Use this data source to query record information of a DNS private zone

Example Usage

Query all records of DNS private zone

```hcl
data "zenlayercloud_zdns_zone_records" "all" {
  zone_id = "<zoneId>"
}
```

Query zone record by record id

```hcl
data "zenlayercloud_zdns_zone_records" "foo" {
  zone_id = "<zoneId>"
  ids     = ["<recordId>"]
}
```

Query zone record by record type

```hcl
data "zenlayercloud_zdns_zone_records" "foo" {
  zone_id     = "<zoneId>"
  record_type = "A"
}
```

Query zone record by record value

```hcl
data "zenlayercloud_zdns_zone_records" "foo" {
  zone_id = "<zoneId>"
  value   = "192.168.0.1"
}
```
