~> **Deprecated** Use `zenlayercloud_zec_vnic` with `stack_type`, `ipv6_internet_charge_type`, `ipv6_bandwidth`, and `rate_limit_mode` instead.

Provides a resource to manage the rate limit mode of an existing public IPv6 address on a vNIC.

This resource does NOT create or delete the public IPv6 address itself; it only adopts an existing public IPv6 (already attached to a vNIC) into Terraform state and allows updates to its `rate_limit_mode`. Destroying this resource only removes it from state.

Example Usage

```hcl
resource "zenlayercloud_zec_vnic_public_ipv6" "demo" {
  nic_id          = "xxxxxxxxxxxxxxxxx"
  rate_limit_mode = "LOOSE"
}
```

Import

vNIC public IPv6 rate limit mode can be imported using the vNIC ID, e.g.

```
$ terraform import zenlayercloud_zec_vnic_public_ipv6.demo nic-id
```
