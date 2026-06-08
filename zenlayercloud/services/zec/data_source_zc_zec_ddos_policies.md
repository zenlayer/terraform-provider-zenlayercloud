Use this data source to query DDoS protection policies.

Example Usage

Query all DDoS policies

```hcl
data "zenlayercloud_zec_ddos_policies" "all" {}
```

Query policies by name (fuzzy match)

```hcl
data "zenlayercloud_zec_ddos_policies" "by_name" {
  policy_name = "prod"
}
```

Query policies by ID

```hcl
data "zenlayercloud_zec_ddos_policies" "by_ids" {
  policy_ids = ["pol-xxxxxxxx", "pol-yyyyyyyy"]
}
```
