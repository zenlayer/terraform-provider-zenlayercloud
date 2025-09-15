Use this data source to query Snapshot policies

Example Usage

Query all snapshots policies

```hcl
data "zenlayercloud_zec_disk_snapshot_policies" "all" {}
```

Query snapshot policies by id

```hcl

data "zenlayercloud_zec_disk_snapshot_policies" "foo" {
  ids = ["<snapshotPolicyId>"]
}
```

Query snapshots by name regex

```hcl
data "zenlayercloud_zec_disk_snapshot_policies" "foo" {
  name_regex = "^example"
}
```

Query snapshots by availability zone

```hcl
data "zenlayercloud_zec_disk_snapshot_policies" "foo" {
  availability_zone = "asia-east-1a"
}
```

