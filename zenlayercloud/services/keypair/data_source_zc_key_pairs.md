Use this data source to query SSH key pair list.

Example Usage

Query all SSH key pair list

```hcl
data "zenlayercloud_key_pairs" "all" {
}
```

Query SSH key pair list by name

```hcl
data "zenlayercloud_key_pairs" "myname" {
  key_name = "myname"
}
```
