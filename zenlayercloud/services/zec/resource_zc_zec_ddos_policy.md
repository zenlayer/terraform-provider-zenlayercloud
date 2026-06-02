Provides a ZEC DDoS protection policy resource.

~> **NOTE:** TCP and UDP cannot be blocked simultaneously in `block_protocol`.

~> **NOTE:** When moving an EIP from one policy to another within the same Terraform config, the target policy must declare `depends_on` the source policy to ensure detach completes before attach.

Example Usage

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

Import

DDoS policies can be imported using the policy ID, e.g.

```
$ terraform import zenlayercloud_zec_ddos_policy.example pol-xxxxxxxx
```
