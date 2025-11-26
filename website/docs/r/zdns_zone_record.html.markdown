---
subcategory: "Zenlayer Private DNS(ZDNS)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zdns_zone_record"
sidebar_current: "docs-zenlayercloud-resource-zdns_zone_record"
description: |-
  Use this resource to create a DNS Private zone record.
---

# zenlayercloud_zdns_zone_record

Use this resource to create a DNS Private zone record.

## Example Usage

1. Create a DNS Private zone

```hcl
resource "zenlayercloud_zdns_zone" "foo" {
  zone_name     = "example.com"
  remark        = "test"
  proxy_pattern = "RECURSION"
}
```

2. Create a DNS Private zone record

```hcl
# Create A type record
resource "zenlayercloud_zdns_zone_record" "foo" {
  zone_id     = zenlayercloud_zdns_zone.foo.id
  record_name = "www"
  type        = "A"
  value       = "192.168.0.11"
  ttl         = 30
  remark      = "Test A Record"
}
```

## Argument Reference

The following arguments are supported:

* `record_name` - (Required, String, ForceNew) The name of the record. such as `www`, `@`.
* `type` - (Required, String, ForceNew) DNS record type. Valid values: 
	- `A`: Maps a domain name to an IP address 
	- `AAAA`: Maps a domain name to an IPv6 address 
	- `CNAME`: Maps a domain name to another domain name 
	- `MX`: Maps a domain name to a mail server address 
	- `TXT`: Text information 
	- `PTR`: Maps an IP address to a domain name for reverse DNS lookup 
	- `SRV`: Specifies servers providing specific services (format: [priority] [weight] [port] [target address], e.g., 0 5 5060 sipserver.example.com).
* `value` - (Required, String) The value of the record.
* `zone_id` - (Required, String, ForceNew) The ID of the private zone.
* `line` - (Optional, String, ForceNew) The resolver line. Default is `default`. Also valid for specified region, such as `asia-east-1`.
* `priority` - (Optional, Int) MX priority, which is required when the record type is `MX`. Range: [1, 99], default: 1.
* `remark` - (Optional, String) Remarks for the record.
* `status` - (Optional, String) Record status. Valid values: `Enabled`, `Disabled`.
* `ttl` - (Optional, Int) The ttl of the Private Zone Record. Measured in second. Range: [5,86400], default: 60.
* `weight` - (Optional, Int) Weight for the record. Only takes effect for type `A` or `AAAA`. Range: [1, 100], default: 1.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the record.


## Import

DNS private zone record can be imported, e.g.

```
$ terraform import zenlayercloud_zdns_zone_record.foo zone-id:record-id
```

