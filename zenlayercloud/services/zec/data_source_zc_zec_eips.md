Use this data source to query zec eip information.

Example Usage

Query all eips

```hcl
data "zenlayercloud_zec_eips" "all" {
}
```

Query eips by region id

```hcl
data "zenlayercloud_zec_eips" "foo" {
  region_id = "asia-east-1"
}
```

Query eips by ids

```hcl
data "zenlayercloud_zec_eips" "foo" {
  ids = ["<eipId>"]
}
```

Query eips by public ip address

```hcl
data "zenlayercloud_zec_eips" "foo" {
	public_ip_address = "128.0.0.1"
}
```

Query eips by name regex

```hcl
data "zenlayercloud_zec_eips" "foo" {
  name_regex = "nginx-ip*"
}
```
