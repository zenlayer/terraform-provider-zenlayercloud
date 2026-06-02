---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_ddos_policies"
sidebar_current: "docs-zenlayercloud-datasource-zec_ddos_policies"
description: |-
  Use this data source to query DDoS protection policies.
---

# zenlayercloud_zec_ddos_policies

Use this data source to query DDoS protection policies.

## Example Usage

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

## Argument Reference

The following arguments are supported:

* `policy_ids` - (Optional, Set: [`String`]) Filter by a list of DDoS policy IDs. Maximum 100.
* `policy_name` - (Optional, String) Filter by policy name. Fuzzy search is supported.
* `result_output_file` - (Optional, String) Used to save results to a local file.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `result` - List of DDoS protection policies.
   * `create_time` - The time when the DDoS policy was created.
   * `id` - The ID of the DDoS policy.
   * `policy_name` - The name of the DDoS policy.
   * `resource_group_id` - The resource group ID the policy belongs to.
   * `resource_group_name` - The resource group name the policy belongs to.
   * `tags` - Tags associated with the policy.


