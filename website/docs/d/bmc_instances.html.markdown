---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_instances"
sidebar_current: "docs-zenlayercloud-datasource-bmc_instances"
description: |-
  Use this data source to query bmc instances.
---

# zenlayercloud_bmc_instances

Use this data source to query bmc instances.

## Example Usage

```hcl
data "zenlayercloud_bmc_instances" "foo" {
  availability_zone = "SEL-A"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) The ID of zone that the bmc instance locates at.
* `hostname` - (Optional, String) The hostname of the instance to be queried.
* `image_id` - (Optional, String) The image of the instance to be queried.
* `instance_ids` - (Optional, Set: [`String`]) IDs of the instances to be queried.
* `instance_name` - (Optional, String) Name of the instances to be queried.
* `instance_status` - (Optional, String) Status of the instances to be queried.
* `instance_type_id` - (Optional, String) Instance type, such as `M6C`.
* `private_ipv4` - (Optional, String) The private ip of the instances to be queried.
* `public_ipv4` - (Optional, String) The public ipv4 of the instances to be queried.
* `resource_group_id` - (Optional, String) The ID of resource group that the instance grouped by.
* `result_output_file` - (Optional, String) Used to save results.
* `subnet_id` - (Optional, String) The ID of vpc subnetwork.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `instance_list` - An information list of instances. Each element contains the following attributes:
   * `availability_zone` - The ID of zone that the bmc instance locates at.
   * `create_time` - Create time of the instance.
   * `expired_time` - Expired time of the instance.
   * `hostname` - The hostname of the instance.
   * `image_id` - The ID of image to use for the instance.
   * `image_name` - The image name to use for the instance.
   * `instance_charge_prepaid_period` - The tenancy (time unit is month) of the prepaid instance.
   * `instance_charge_type` - The charge type of instance.
   * `instance_id` - ID of the instances.
   * `instance_name` - The name of the instance.
   * `instance_status` - Current status of the instance.
   * `instance_type_id` - The type of the instance.
   * `internet_max_bandwidth_out` - Maximum outgoing bandwidth to the public network, measured in Mbps (Mega bits per second).
   * `key_id` - The ssh key pair id used for the instance.
   * `nic_lan_name` - The lan name of the nic. The lan name should be a combination of 4 to 10 characters comprised of letters (case insensitive), numbers. The lan name must start with letter. Modifying will cause the instance reset.
   * `nic_wan_name` - The wan name of the nic. The wan name should be a combination of 4 to 10 characters comprised of letters (case insensitive), numbers. The wan name must start with letter. Modifying will cause the instance reset.
   * `partitions` - Partition for the instance. Modifying will cause the instance reset.
      * `fs_path` - The drive letter(windows) or device name(linux) for the partition.
      * `fs_type` - The type of the partitioned file.
      * `size` - The size of the partitioned disk.
   * `private_ipv4_addresses` - Private Ipv4 addresses of the instance.
   * `public_ipv4_addresses` - Public Ipv4 addresses of the instance.
   * `public_ipv6_addresses` - Public Ipv6 addresses of the instance.
   * `raid_config_custom` - Custom config for instance raid. Modifying will cause the instance reset.
      * `disk_sequence` - The sequence of disk to make raid.
      * `raid_type` - Simple config for raid.
   * `raid_config_type` - Simple config for instance raid. Modifying will cause the instance reset.
   * `resource_group_id` - The ID of resource group that the instance belongs to.
   * `resource_group_name` - The name of resource group that the instance belongs to.
   * `traffic_package_size` - Traffic package size.


