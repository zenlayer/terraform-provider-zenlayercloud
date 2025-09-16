Provides a resource to attached ZEC disk to an instance.

Example Usage

```hcl
resource "zenlayercloud_zec_disk_attachment" "test" {
	disk_id     = "<diskId>"
	instance_id = "<instanceId>"
}
```

Import

Disk attachment can be imported, e.g.

```
$ terraform import zenlayercloud_zec_disk_attachment.test disk-id:instance-id
```
