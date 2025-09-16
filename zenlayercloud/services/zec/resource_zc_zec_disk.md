Provide a resource to create data disk.

Example Usage

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

Import

Disk instance can be imported, e.g.

```
$ terraform import zenlayercloud_zec_disk.test disk-id
```
