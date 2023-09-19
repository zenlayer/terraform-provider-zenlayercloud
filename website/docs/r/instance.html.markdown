---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_instance"
sidebar_current: "docs-zenlayercloud-resource-instance"
description: |-
  Provides a instance resource.
---

# zenlayercloud_instance

Provides a instance resource.

~> **NOTE:** You can launch an instance for a private network via specifying parameter `subnet_id`.

~> **NOTE:** At present, 'PREPAID' instance cannot be deleted and must wait it to be outdated and released automatically.

## Example Usage

```hcl
data "zenlayercloud_zones" "default" {

}

data "zenlayercloud_instance_types" "default" {
  availability_zone = data.zenlayercloud_zones.default.zones.0.id
}

# Get a centos image which also supported to install on given instance type
data "zenlayercloud_images" "default" {
  availability_zone = data.zenlayercloud_zones.default.zones.0.id
  category          = "CentOS"
}

resource "zenlayercloud_subnet" "default" {
  name       = "test-subnet"
  cidr_block = "10.0.10.0/24"
}

# Create a web server
resource "zenlayercloud_instance" "web" {
  availability_zone    = data.zenlayercloud_zones.default.zones.0.id
  image_id             = data.zenlayercloud_images.default.images.0.image_id
  internet_charge_type = "ByBandwidth"
  instance_type        = data.zenlayercloud_instance_types.default.instance_types.0.id
  password             = "Example~123"
  instance_name        = "web"
  subnet_id            = zenlayercloud_subnet.default.id
  system_disk_size     = 100
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The ID of zone that the instance locates at.
* `instance_type` - (Required, String, ForceNew) The type of the instance.
* `internet_charge_type` - (Required, String, ForceNew) Internet charge type of the instance, Valid values are `ByBandwidth`, `ByTrafficPackage`, `ByInstanceBandwidth95` and `ByClusterBandwidth95`. This value currently not support to change.
* `system_disk_size` - (Required, Int, ForceNew) Size of the system disk. unit is GB. If modified, the instance may force stop.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the instance. Default is `false`. If set true, the instance will be permanently deleted instead of being moved into the recycle bin.
* `image_id` - (Optional, String) The image to use for the instance. Changing `image_id` will cause the instance reset.
* `instance_charge_prepaid_period` - (Optional, Int) The tenancy (time unit is month) of the prepaid instance, NOTE: it only works when instance_charge_type is set to `PREPAID`.
* `instance_charge_type` - (Optional, String, ForceNew) The charge type of instance. Valid values are `PREPAID`, `POSTPAID`. The default is `POSTPAID`. Note: `PREPAID` instance may not allow to delete before expired.
* `instance_name` - (Optional, String) The name of the instance. The max length of instance_name is 64, and default value is `Terraform-Instance`.
* `internet_max_bandwidth_out` - (Optional, Int) Maximum outgoing bandwidth to the public network, measured in Mbps (Mega bits per second).
* `key_id` - (Optional, String) The key pair id to use for the instance. Changing `key_id` will cause the instance reset.
* `password` - (Optional, String) Password for the instance. The max length of password is 16.
* `resource_group_id` - (Optional, String) The resource group id the instance belongs to, default to Default Resource Group.
* `subnet_id` - (Optional, String, ForceNew) The ID of a VPC subnet. If you want to create instances in a VPC network, this parameter must be set.
* `traffic_package_size` - (Optional, Float64) Traffic package size. Only valid when the charge type of instance is `ByTrafficPackage` and the instance charge type is `PREPAID`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the instance.
* `expired_time` - Expired time of the instance.
* `image_name` - The image name to use for the instance.
* `instance_status` - Current status of the instance.
* `private_ip_addresses` - Private Ip addresses of the instance.
* `public_ip_addresses` - Public Ip addresses of the instance.
* `resource_group_name` - The resource group name the instance belongs to, default to Default Resource Group.


## Import

Instance can be imported using the id, e.g.

```
terraform import zenlayercloud_instance.foo 123123xxx
```

