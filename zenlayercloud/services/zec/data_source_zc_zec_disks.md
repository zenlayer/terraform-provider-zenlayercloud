Use this data source to query zec disk information.

Example Usage

Query all disks storages

```hcl
data "zenlayercloud_zec_disks" "all" {
}
```

Query disks by availability zone

```hcl
data "zenlayercloud_zec_disks" "zone_disk" {
  availability_zone = "asia-east-1"
}
```

Query disks by ids

```hcl
data "zenlayercloud_zec_disks" "zone_disk" {
  ids = ["<diskId>"]
}
```

Query disks by disk type

```hcl
data "zenlayercloud_zec_disks" "system_disk" {
  disk_type = "SYSTEM"
}
```

Query disks by attached instance

```hcl
data "zenlayercloud_zec_disks" "instance_disk" {
  instance_id = "<instanceId>"
}
```

Query disks by name regex

```hcl
data "zenlayercloud_zec_disks" "name_disk" {
  name_regex = "disk20*"
}
```
