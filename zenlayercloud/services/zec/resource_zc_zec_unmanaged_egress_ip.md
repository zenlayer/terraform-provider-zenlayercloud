Provides a resource to manage the rate limit mode of an existing unmanaged egress IP.

This resource does NOT create or delete the unmanaged egress IP itself; it only adopts an existing unmanaged egress IP into Terraform state and allows updates to its `rate_limit_mode`. Destroying this resource only removes it from state.

Example Usage

```hcl
resource "zenlayercloud_zec_unmanaged_egress_ip" "demo" {
  unmanaged_egress_ip_id = "xxxxxxxxxxxxxxxxx"
  rate_limit_mode        = "LOOSE"
}
```

Import

Unmanaged egress IP rate limit mode can be imported, e.g.

```
$ terraform import zenlayercloud_zec_unmanaged_egress_ip.demo unmanaged-egress-ip-id
```
