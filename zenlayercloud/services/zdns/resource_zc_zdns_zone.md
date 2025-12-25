Use this resource to create a DNS Private zone

For more information about Zenlayer DNS, see the Zenlayer Documentation on [ZDNS Service](https://docs.console.zenlayer.com/welcome/elastic-compute/overview/zdns-service)

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
