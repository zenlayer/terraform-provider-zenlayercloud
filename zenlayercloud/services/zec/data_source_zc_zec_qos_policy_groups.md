Use this data source to query ZEC QoS policy group information.

Example Usage

Query all QoS policy groups

```hcl
data "zenlayercloud_zec_qos_policy_groups" "all" {
}
```

Query QoS policy groups by region ID

```hcl
data "zenlayercloud_zec_qos_policy_groups" "by_region" {
  region_id = "asia-southeast-1"
}
```

Query QoS policy groups by IDs

```hcl
data "zenlayercloud_zec_qos_policy_groups" "by_ids" {
  ids = ["<qosPolicyGroupId>"]
}
```

Query QoS policy groups by name regex

```hcl
data "zenlayercloud_zec_qos_policy_groups" "by_name" {
  name_regex = "^example-*"
}
```

Query the QoS policy group that a specific resource belongs to

```hcl
data "zenlayercloud_zec_qos_policy_groups" "by_member" {
  resource_id = "<eipId>"
}
```
