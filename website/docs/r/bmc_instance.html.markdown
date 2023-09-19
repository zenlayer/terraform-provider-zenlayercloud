---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_instance"
sidebar_current: "docs-zenlayercloud-resource-bmc_instance"
description: |-
  Provides a BMC instance resource.
---

# zenlayercloud_bmc_instance

Provides a BMC instance resource.

~> **NOTE:** You can launch an BMC instance for a private network via specifying parameter `subnet_id`.

~> **NOTE:** At present, 'PREPAID' instance cannot be deleted and must wait it to be outdated and released automatically.

## Example Usage

```hcl
data "zenlayercloud_bmc_zones" "default" {

}

data "zenlayercloud_bmc_instance_types" "default" {
  availability_zone = data.zenlayercloud_bmc_zones.default.zones.0.id
}

# Get a centos image which also supported to install on given instance type
data "zenlayercloud_bmc_images" "default" {
  catalog          = "centos"
  instance_type_id = data.zenlayercloud_bmc_instance_types.default.instance_types.0.id
}

resource "zenlayercloud_bmc_subnet" "default" {
  availability_zone = data.zenlayercloud_bmc_zones.default.zones.0.id
  name              = "test-subnet"
  cidr_block        = "10.0.10.0/24"
}

# Create a web server
resource "zenlayercloud_bmc_instance" "web" {
  availability_zone    = data.zenlayercloud_bmc_zones.default.zones.0.id
  image_id             = data.zenlayercloud_bmc_images.default.images.0.image_id
  internet_charge_type = "ByBandwidth"
  instance_type_id     = data.zenlayercloud_bmc_instance_types.default.instance_types.0.id
  password             = "Example~123"
  instance_name        = "web"
  subnet_id            = zenlayercloud_bmc_subnet.default.id
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The ID of zone that the bmc instance locates at.
* `instance_type_id` - (Required, String, ForceNew) The type of the instance.
* `internet_charge_type` - (Required, String, ForceNew) Internet charge type of the instance, Valid values are `ByBandwidth`, `ByTrafficPackage`, `ByInstanceBandwidth95` and `ByClusterBandwidth95`. This value currently not support to change.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the instance. Default is `false`. If set true, the instance will be permanently deleted instead of being moved into the recycle bin.
* `hostname` - (Optional, String) The hostname of the instance. The name should be a combination of 2 to 64 characters comprised of letters (case insensitive), numbers, hyphens (-) and Period (.), and the name must be start with letter. The default value is `Terraform-Instance`. Modifying will cause the instance reset.
* `image_id` - (Optional, String) The image to use for the instance. Changing `image_id` will cause the instance reset.
* `instance_charge_prepaid_period` - (Optional, Int) The tenancy (time unit is month) of the prepaid instance, NOTE: it only works when instance_charge_type is set to `PREPAID`.
* `instance_charge_type` - (Optional, String, ForceNew) The charge type of instance. Valid values are `PREPAID`, `POSTPAID`. The default is `POSTPAID`. Note: `PREPAID` instance may not allow to delete before expired.
* `instance_name` - (Optional, String) The name of the instance. The max length of instance_name is 64, and default value is `Terraform-Instance`.
* `internet_max_bandwidth_out` - (Optional, Int) Maximum outgoing bandwidth to the public network, measured in Mbps (Mega bits per second).
* `nic_lan_name` - (Optional, String) The lan name of the nic. The lan name should be a combination of 4 to 10 characters comprised of letters (case insensitive), numbers. The lan name must start with letter. Modifying will cause the instance reset.
* `nic_wan_name` - (Optional, String) The wan name of the nic. The wan name should be a combination of 4 to 10 characters comprised of letters (case insensitive), numbers. The wan name must start with letter. Modifying will cause the instance reset.
* `partitions` - (Optional, List) Partition for the instance. Modifying will cause the instance reset.
* `password` - (Optional, String) Password for the instance. The max length of password is 16. Modifying will cause the instance reset.
* `raid_config_custom` - (Optional, List) Custom config for instance raid. Modifying will cause the instance reset.
* `raid_config_type` - (Optional, String) Simple config for instance raid. Modifying will cause the instance reset.
* `resource_group_id` - (Optional, String) The resource group id the instance belongs to, default to Default Resource Group.
* `ssh_keys` - (Optional, Set: [`String`]) The ssh keys to use for the instance. The max number of ssh keys is 5. Modifying will cause the instance reset.
* `subnet_id` - (Optional, String) The ID of a VPC subnet. If you want to create instances in a VPC network, this parameter must be set.
* `traffic_package_size` - (Optional, Float64) Traffic package size. Only valid when the charge type of instance is `ByTrafficPackage` and the instance charge type is `PREPAID`.

The `partitions` object supports the following:

* `fs_path` - (Required, String) The drive letter(windows) or device name(linux) for the partition.
* `fs_type` - (Required, String) The type of the partitioned file.
* `size` - (Required, Int) The size of the partitioned disk.

The `raid_config_custom` object supports the following:

* `disk_sequence` - (Required, List) The sequence of disk to make raid.
* `raid_type` - (Required, String) Simple config for raid.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the instance.
* `expired_time` - Expired time of the instance.
* `image_name` - The image name to use for the instance.
* `instance_status` - Current status of the instance.
* `primary_ipv4_address` - Primary Ipv4 address of the instance.
* `private_ip_addresses` - Private Ip addresses of the instance.
* `public_ipv4_addresses` - Public Ipv4 addresses bind to the instance.
* `public_ipv6_addresses` - Public Ipv6 addresses of the instance.
* `resource_group_name` - The resource group name the instance belongs to, default to Default Resource Group.


## Import

BMC instance can be imported using the id, e.g.

```
terraform import zenlayercloud_bmc_instance.foo 123123xxx
```

