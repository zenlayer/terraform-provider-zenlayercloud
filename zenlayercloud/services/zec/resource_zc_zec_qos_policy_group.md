Provide a resource to manage a QoS policy group. A QoS policy group enforces a shared bandwidth limit across its member IPs (EIP, IPv6, or UNMANAGED egress IP).

Use `zenlayercloud_zec_qos_policy_group_member` to add members to the group.

Example Usage

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

Import

QoS policy group can be imported, e.g.

```
$ terraform import zenlayercloud_zec_qos_policy_group.example <qosPolicyGroupId>
```
