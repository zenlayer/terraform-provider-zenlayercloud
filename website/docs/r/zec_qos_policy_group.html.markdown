---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_qos_policy_group"
sidebar_current: "docs-zenlayercloud-resource-zec_qos_policy_group"
description: |-
  Provide a resource to manage a QoS policy group. A QoS policy group enforces a shared bandwidth limit across its member IPs (EIP, IPv6, or UNMANAGED egress IP).
---

# zenlayercloud_zec_qos_policy_group

Provide a resource to manage a QoS policy group. A QoS policy group enforces a shared bandwidth limit across its member IPs (EIP, IPv6, or UNMANAGED egress IP).

Use `zenlayercloud_zec_qos_policy_group_member` to add members to the group.

## Example Usage

```hcl
variable "region" {
  default = "asia-southeast-1"
}

resource "zenlayercloud_zec_qos_policy_group" "example" {
  region_id       = var.region
  name            = "example-qos-group"
  bandwidth_limit = 100
  rate_limit_mode = "LOOSE"
  tags = {
    "env" = "test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `bandwidth_limit` - (Required, Int) The shared bandwidth limit in Mbps.
* `name` - (Required, String) The name of the QoS policy group.
* `region_id` - (Required, String, ForceNew) The region ID where the QoS policy group is located.
* `rate_limit_mode` - (Optional, String) The rate limit mode of the QoS policy group. Default is LOOSE, Valid values: `LOOSE` - each forwarding server starts with the full group cap, allowing a single connection to reach maximum speed immediately but may briefly exceed the cap under concurrent traffic; `STRICT` - bandwidth is divided evenly across forwarding servers so the group cap is never exceeded, but multiple parallel flows are needed to fully utilize the cap.
* `resource_group_id` - (Optional, String) The resource group ID the QoS policy group belongs to. Defaults to the default resource group.
* `tags` - (Optional, Map) The tags of the QoS policy group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - The creation time of the QoS policy group.
* `member_count` - The number of members currently in the QoS policy group.
* `resource_group_name` - The resource group name the QoS policy group belongs to.


## Import

QoS policy group can be imported, e.g.

```
$ terraform import zenlayercloud_zec_qos_policy_group.example <qosPolicyGroupId>
```

