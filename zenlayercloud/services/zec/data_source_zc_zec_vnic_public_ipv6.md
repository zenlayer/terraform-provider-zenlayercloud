Use this data source to query the public IPv6 attached to a single vNIC.

Example Usage

```hcl
data "zenlayercloud_zec_vnic_public_ipv6" "demo" {
  nic_id = "1680855999352675875"
}

output "ipv6_rate_limit_mode" {
  value = data.zenlayercloud_zec_vnic_public_ipv6.demo.rate_limit_mode
}
```
