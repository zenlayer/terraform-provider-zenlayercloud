Use this data source to query public CIDR blocks.

Example Usage

Query all public CIDR blocks

```hcl
data "zenlayercloud_zec_cidrs" "all" {}
```

Query CIDRs by id

```hcl
data "zenlayercloud_zec_cidrs" "snapshot" {
  ids = ["<cidrId>"] 
}
```

Query CIDRs by name regex

```hcl
data "zenlayercloud_zec_cidrs" "foo" {
  name_regex = "^example"
}
```

Query CIDRs by region id

```hcl
data "zenlayercloud_zec_cidrs" "foo" {
	region_id = "asia-east-1"
}
```
