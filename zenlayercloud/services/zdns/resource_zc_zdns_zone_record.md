Use this resource to create a DNS Private zone record.

Example Usage

1. Create a DNS Private zone

```hcl
resource "zenlayercloud_zdns_zone" "foo" {
	zone_name = "example.com"
	remark = "test"
	proxy_pattern = "RECURSION"
}
```

2. Create a DNS Private zone record
```hcl
# Create A type record
resource "zenlayercloud_zdns_zone_record" "foo" {
  zone_id = zenlayercloud_zdns_zone.foo.id
  record_name = "www"
  type = "A"
  value = "192.168.0.11"
  ttl = 30
  remark = "Test A Record"
}
```

Import

DNS private zone record can be imported, e.g.

```
$ terraform import zenlayercloud_zdns_zone_record.foo zone-id:record-id
```
