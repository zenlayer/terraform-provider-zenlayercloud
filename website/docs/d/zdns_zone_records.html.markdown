---
subcategory: "Zenlayer Private DNS(zdns)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zdns_zone_records"
sidebar_current: "docs-zenlayercloud-datasource-zdns_zone_records"
description: |-
  Use this data source to query record information of a DNS private zone
---

# zenlayercloud_zdns_zone_records

Use this data source to query record information of a DNS private zone

## Example Usage

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

## Argument Reference

The following arguments are supported:

* `zone_id` - (Required, String) ID of the DNS private zone.
* `ids` - (Optional, Set: [`String`]) IDs of the records to be queried.
* `record_type` - (Optional, String) Type of the records to be queried.
* `record_value` - (Optional, String) Value of the records to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `records` - An information list of private DNS records. Each element contains the following attributes:
   * `create_time` - Creation time of the private DNS record.
   * `id` - ID of the private DNS record.
   * `priority` - Priority of the private DNS record.
   * `record_name` - Name of the private DNS record.
   * `record_type` - Type of the private DNS record.
   * `record_value` - Value of the private DNS record.
   * `status` - Status of the private DNS record.
   * `ttl` - TTL of the private DNS record.
   * `weight` - Weight of the private DNS record.


