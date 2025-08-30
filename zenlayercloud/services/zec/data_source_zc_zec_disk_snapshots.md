Use this data source to query Snapshots

Example Usage

Query all snapshots

```hcl
data "zenlayercloud_zec_disk_snapshots" "all" {}
```

Query snapshots by id

```hcl
# Create a snapshot
resource "zenlayercloud_zec_snapshot" "snapshot" {
  disk_id = "<diskId>"
  name    = "example-snapshot"
}

# Query snapshots using data source
data "zenlayercloud_zec_disk_snapshots" "foo" {
  ids = [zenlayercloud_zec_snapshot.snapshot.id]
}
```

Query snapshots by name regex

```hcl
data "zenlayercloud_zec_disk_snapshots" "foo" {
  name_regex = "^example"
}
```

Query snapshots by availability zone

```hcl
data "zenlayercloud_zec_disk_snapshots" "foo" {
  availability_zone = "asia-east-1a"
}
```

Query snapshots by snapshot type

```hcl
data "zenlayercloud_zec_disk_snapshots" "foo" {
  snapshot_type = "Auto"
}
```

Query snapshots by disk id

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

resource "zenlayercloud_zec_snapshot" "snapshot" {
  disk_id = "<diskId>"
  name    = "example-snapshot"
}

data "zenlayercloud_zec_disk_snapshots" "foo" {
  disk_id = zenlayercloud_zec_snapshot.snapshot.disk_id
}
```

