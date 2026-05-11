Use this data source to query public IPv6 CIDR blocks.

Example Usage

Query all public IPv6 CIDR blocks

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "all" {}
```

Query IPv6 CIDRs by id

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "foo" {
  ids = ["<cidrId>"]
}
```

Query IPv6 CIDRs by name regex

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "foo" {
  name_regex = "^example"
}
```

Query IPv6 CIDRs by region id

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "foo" {
  region_id = "asia-east-1"
}
```

Query BYOIP IPv6 CIDRs by ASN

```hcl
data "zenlayercloud_zec_ipv6_cidrs" "foo" {
  asn = 62210
}
```
