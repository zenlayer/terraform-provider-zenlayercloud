Use this resource to manage DNS private zone with zec VPCs attachment

~> **NOTE:** The current resource is used to manage all the zec VPCs of a DNS private zone.

Example Usage

1. Prepare a DNS Private zone and a zec VPC

```hcl
resource "zenlayercloud_zdns_zone" "foo" {
	zone_name = "example.com"
	remark = "test"
	proxy_pattern = "RECURSION"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name = "example"
  cidr_block = "10.0.0.0/24"
  enable_ipv6 = true
  mtu = 1300
}
```

2. Bind the VPC to DNS Private zone
```hcl
resource "zenlayercloud_zdns_zone_vpc_set_attachment" "foo" {
  zone_id = zenlayercloud_zdns_zone.foo.id
  vpc_ids = [zenlayercloud_zec_vpc.foo.id]
}
```

Import

DNS private zone vpc attachment can be imported, e.g.

```
$ terraform import zenlayercloud_zdns_zone_vpc_set_attachment.foo zone-id
```
