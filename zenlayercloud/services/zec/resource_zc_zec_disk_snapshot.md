Provides a resource to create snapshot for ZEC disk.

Example Usage

Prepare a disk

```hcl

variable "availability_zone" {
  default = "asia-east-1a"
}

resource "zenlayercloud_zec_disk" "test" {
  availability_zone = var.availability_zone
  disk_name         = "Disk-20G"
  disk_size         = 60
  disk_category     = "Standard NVMe SSD"
}
```

Create a snapshot

```hcl
resource "zenlayercloud_zec_disk_snapshot" "snapshot" {
  disk_id = zenlayercloud_zec_disk.test.id
  name    = "example-snapshot"
}
```

Import

Snapshot can be imported, e.g.

```
$ terraform import zenlayercloud_zec_disk_snapshot.snapshot snapshot-id
```
