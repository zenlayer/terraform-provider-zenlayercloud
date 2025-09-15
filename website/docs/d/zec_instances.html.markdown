---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_instances"
sidebar_current: "docs-zenlayercloud-datasource-zec_instances"
description: |-
  Use this data source to query ZEC instances.
---

# zenlayercloud_zec_instances

Use this data source to query ZEC instances.

## Example Usage

Query all instances

```hcl
data "zenlayercloud_zec_instances" "foo" {
}
```

Query zec instances by ids

```hcl
data "zenlayercloud_zec_instances" "foo" {
  ids = ["<instanceId>"]
}
```

Query zec instances by availability zone

```hcl
data "zenlayercloud_zec_instances" "foo" {
  availability_zone = "asia-southeast-1a"
}
```

Query zec instances by name regex

```hcl
data "zenlayercloud_zec_instances" "foo" {
  name_regex = "test*"
}
```

Query zec instances by image id

```hcl
data "zenlayercloud_zec_instances" "foo" {
  image_id = "<imageId>"
}
```

Query zec instances by IPv4 address (including private & public IPv4)

```hcl
data "zenlayercloud_zec_instances" "foo" {
  ipv4_address = "10.0.0.2"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) The ID of zone that the bmc instance locates at.
* `ids` - (Optional, Set: [`String`]) IDs of the ZEC instances to be queried.
* `image_id` - (Optional, String) The image of the ZEC instance to be queried.
* `instance_status` - (Optional, String) Status of the ZEC instances to be queried.
* `ipv4_address` - (Optional, String) The ipv4 address of the ZEC instances to be queried.
* `ipv6_address` - (Optional, String) The ipv6 address of the ZEC instances to be queried.
* `name_regex` - (Optional, String) A regex string to filter results by instance name.
* `resource_group_id` - (Optional, String) The ID of resource group that the ZEC instance grouped by.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instances` - An information list of instances. Each element contains the following attributes:
   * `availability_zone` - The ID of zone that the ZEC instance locates at.
   * `cpu` - The number of CPU cores of the ZEC instance.
   * `create_time` - Create time of the ZEC instance.
   * `data_disks` - List of data disk. Each element contains the following attributes:
      * `data_disk_category` - Category of the data disk.
      * `data_disk_id` - Image ID of the data disk.
      * `data_disk_size` - Size of the data disk.
   * `id` - ID of the ZEC instances.
   * `image_id` - The ID of image to use for the ZEC instance.
   * `image_name` - The image name to use for the ZEC instance.
   * `instance_name` - The name of the ZEC instance.
   * `instance_status` - Current status of the ZEC instance.
   * `instance_type` - The type of the ZEC instance.
   * `memory` - Memory capacity of the ZEC instance, unit in GiB.
   * `nic_network_type` - The Network card mode for the ZEC instance. Valid values: FailOver,VirtioOnly,VfOnly.
   * `private_ip_addresses` - Public Ipv4 addresses of the ZEC instance.
   * `public_ip_addresses` - Public Ipv6 addresses of the ZEC instance.
   * `resource_group_id` - The ID of resource group that the ZEC instance belongs to.
   * `resource_group_name` - The name of resource group that the ZEC instance belongs to.
   * `system_disk_category` - Category of the system disk.
   * `system_disk_id` - ID of the system disk.
   * `system_disk_size` - Size of the system disk.


