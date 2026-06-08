---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_ddos_policy"
sidebar_current: "docs-zenlayercloud-resource-zec_ddos_policy"
description: |-
  Provides a ZEC DDoS protection policy resource.
---

# zenlayercloud_zec_ddos_policy

Provides a ZEC DDoS protection policy resource.

~> **NOTE:** TCP and UDP cannot be blocked simultaneously in `block_protocol`.

~> **NOTE:** When moving an EIP from one policy to another within the same Terraform config, the target policy must declare `depends_on` the source policy to ensure detach completes before attach.

## Example Usage

Create a basic DDoS policy

```hcl
resource "zenlayercloud_zec_ddos_policy" "basic" {
  policy_name = "my-ddos-policy"
}
```

Create a policy with EIP binding and IP blacklist

```hcl
resource "zenlayercloud_zec_ddos_policy" "example" {
  policy_name  = "my-ddos-policy"
  ipv4_id_list = [zenlayercloud_zec_eip.example.id]

  black_ip_list    = ["1.2.3.4"]
  ip_black_timeout = 60

  block_protocol = ["ICMP"]
  block_regions  = ["CN"]
}
```

Create a policy with port blocking and traffic control

```hcl
resource "zenlayercloud_zec_ddos_policy" "full" {
  policy_name = "full-ddos-policy"

  port {
    protocol       = "UDP"
    src_port_start = 0
    src_port_end   = 65535
    dst_port_start = 53
    dst_port_end   = 53
    action         = "Drop"
  }

  traffic_control {
    bps_enabled = true
    bps         = 104857600
    pps_enabled = true
    pps         = 10000
  }

  tags = {
    env = "prod"
  }
}
```

## Argument Reference

The following arguments are supported:

* `policy_name` - (Required, String) Name of the DDoS protection policy. 2-63 characters, only letters, digits, `-`, and `.` are allowed, must start and end with a letter or digit.
* `black_ip_list` - (Optional, Set: [`String`]) List of blacklisted IP addresses.
* `block_protocol` - (Optional, Set: [`String`]) List of protocols to block. Valid values: `TCP`, `UDP`, `ICMP`. Note: `TCP` and `UDP` cannot be blocked simultaneously.
* `block_regions` - (Optional, Set: [`String`]) List of region IDs to block. Use `DescribePolicyRegions` API to get available region IDs.
* `fingerprint_rule` - (Optional, List) Fingerprint filtering rules.
* `ip_black_timeout` - (Optional, Int) Blacklist timeout in minutes. Valid range: 1-10080. Required when black_ip_list is set.
* `ipv4_id_list` - (Optional, Set: [`String`]) List of EIP IDs to attach to this policy. Each EIP can only be attached to one policy at a time. When moving an EIP from one policy to another, the target policy resource must declare `depends_on` the source policy to ensure detach completes before attach.
* `port` - (Optional, List) Port blocking rules.
* `reflect_udp_port` - (Optional, List) Additional UDP reflection attack source ports to block, on top of the system built-in defaults. Use `DescribeReflectUdpPortOptions` to query the built-in default ports.
* `resource_group_id` - (Optional, String) The resource group ID. If not specified, the default resource group is used.
* `tags` - (Optional, Map) Tags associated with the DDoS policy.
* `traffic_control` - (Optional, List) Source IP rate limiting configuration.
* `white_ip_list` - (Optional, Set: [`String`]) List of whitelisted IP addresses.

The `fingerprint_rule` object supports the following:

* `action` - (Required, String) Action to take on match. Valid values: `Drop`.
* `dst_port_end` - (Required, Int) Destination port range end value. Range: 0-65535.
* `dst_port_start` - (Required, Int) Destination port range start value. Range: 0-65535.
* `match_bytes` - (Required, String) Bytes to match in the payload. Hexadecimal lowercase, zero-padded to 2 digits (e.g. `deadbeef`).
* `max_pkt_length` - (Required, Int) Maximum packet length to filter. Range: 1-1500.
* `min_pkt_length` - (Required, Int) Minimum packet length to filter. Range: 1-1500.
* `protocol` - (Required, String) Protocol type. Valid values: `TCP`, `UDP`, `ICMP`.
* `src_port_end` - (Required, Int) Source port range end value. Range: 0-65535.
* `src_port_start` - (Required, Int) Source port range start value. Range: 0-65535.
* `offset` - (Optional, Int) Payload offset for fingerprint matching. Range: 0-1500.

The `port` object supports the following:

* `action` - (Required, String) Action to take on match. Valid values: `Drop`.
* `dst_port_end` - (Required, Int) Destination port range end value. Range: 0-65535.
* `dst_port_start` - (Required, Int) Destination port range start value. Range: 0-65535.
* `protocol` - (Required, String) Protocol type. Valid values: `TCP`, `UDP`.
* `src_port_end` - (Required, Int) Source port range end value. Range: 0-65535.
* `src_port_start` - (Required, Int) Source port range start value. Range: 0-65535.

The `reflect_udp_port` object supports the following:

* `port` - (Required, Int) UDP reflection source port to block. Range: 0-65535.

The `traffic_control` object supports the following:

* `bps_enabled` - (Optional, Bool) Whether to enable bps rate limiting.
* `bps` - (Optional, Int) Bps rate limit value. Valid range: [8192, 2147483648].
* `pps_enabled` - (Optional, Bool) Whether to enable pps rate limiting.
* `pps` - (Optional, Int) Pps rate limit value. Valid range: [32, 50000].
* `syn_bps_enabled` - (Optional, Bool) Whether to enable SYN bps rate limiting.
* `syn_bps` - (Optional, Int) SYN bps rate limit value. Valid range: [8192, 2147483648].
* `syn_pps_enabled` - (Optional, Bool) Whether to enable SYN pps rate limiting.
* `syn_pps` - (Optional, Int) SYN pps rate limit value. Valid range: [1, 100000].

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - The time when the DDoS policy was created.


## Import

DDoS policies can be imported using the policy ID, e.g.

```
$ terraform import zenlayercloud_zec_ddos_policy.example pol-xxxxxxxx
```

