Provides a resource to create auto snapshot policy

Example Usage

```hcl
resource "zenlayercloud_zec_disk_snapshot_policy" "example" {
  availability_zone = "asia-east-1a"
  name             = "example-snapshot-policy"
  repeat_week_days = [1]
  hours            = [12]
  retention_days   = 7
}
```

Import

Snapshot Policy can be imported using the `id`, e.g.

```bash
$ terraform import zc_zec_disk_snapshot_policy.example policy-id
```
