Provides a resource to attached ZEC disk to an auto snapshot policy.

Example Usage

```hcl
var "availability_zone" {
  default = "asia-east-1a"
}

resource "zenlayercloud_zec_disk" "test" {
  availability_zone = var.availability_zone
  disk_name         = "Disk-20G"
  disk_size         = 60
  disk_category     = "Standard NVMe SSD"
}

resource "zenlayercloud_zec_disk_snapshot_policy" "test" {
  availability_zone = var.availability_zone
  name              = "example-snapshot-policy"
  repeat_week_days  = [1]
  hours             = [12]
  retention_days    = 7
}

resource "zenlayercloud_zec_disk_snapshot_policy_attachment" "test" {
  disk_id                 = zenlayercloud_zec_disk.test.id
  auto_snapshot_policy_id = zenlayercloud_zec_disk_snapshot_policy.test.id
}
```

Import

Disk Snapshot Policy attachment can be imported, e.g.

```
$ terraform import zenlayercloud_zec_disk_snapshot_policy_attachment.test disk-id
```
