Use this resource to create a DNS Private zone

Example Usage

Create a DNS Private zone

```hcl
resource "zenlayercloud_zdns_zone" "foo" {
	zone_name = "example.com"
	remark = "test"
	proxy_pattern = "RECURSION"
}
```

Import

DNS private zone can be imported, e.g.

```
$ terraform import zenlayercloud_zdns_zone.foo zone-id
```
