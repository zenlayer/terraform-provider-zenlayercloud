---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_instance"
sidebar_current: "docs-zenlayercloud-resource-zec_instance"
description: |-
  Provides a instance resource.
---

# zenlayercloud_zec_instance

Provides a instance resource.

~> **NOTE:** Currently it's not able to create default public IPv4 through instance resource. Please use `zenlayercloud_zec_eip` resource to create EIP and then call resource `zenlayer_zec_eip_association` to bind it to resource

~> **NOTE:** Currently this resource doesn't support create instance through `Windows` and `Generic` image.

## Example Usage

```hcl
variable "availability_zone" {
  default = "asia-east-1a"
}

variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_security_group" "sg" {
  name = "Test-SecurityGroup1"
}

resource "zenlayercloud_zec_vpc_security_group_attachment" "foo" {
  vpc_id            = zenlayercloud_zec_vpc.foo.id
  security_group_id = zenlayercloud_zec_security_group.sg.id
}

# Create subnet (IPv4 IP stack)
resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "test-subnet"
  cidr_block = "10.0.0.0/24"
}

# Create instance Using key pair
data "zenlayercloud_key_pairs" "all" {
}

data "zenlayercloud_zec_images" "ubuntu" {
  availability_zone = var.availability_zone
  category          = "Ubuntu"
}

# Create a Instance
resource "zenlayercloud_zec_instance" "instance" {
  availability_zone = var.availability_zone
  instance_type     = "z2a.cpu.1"
  image_id          = data.zenlayercloud_zec_images.ubuntu.images.0.id
  instance_name     = "Example-Instance"
  key_id            = data.zenlayercloud_key_pairs.all.key_pairs.0.key_id
  subnet_id         = zenlayercloud_zec_subnet.ipv4.id
  system_disk_size  = 20
}
```

# Import

Instance can be imported using the id, e.g.

```hcl
terraform import zenlayercloud_zec_instance.instance instance-id
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required, String, ForceNew) The ID of zone that the ZEC instance locates at. such as `asia-southeast-1a`.
* `image_id` - (Required, String) The image to use for the ZEC instance. Changing `image_id` will cause the ZEC instance reset.
* `instance_type` - (Required, String, ForceNew) The type of the ZEC instance. such as `z2a.cpu.4`.
* `subnet_id` - (Required, String, ForceNew) The ID of a VPC subnet. Note: The **IPv6 only** stack subnet is not support for instance creation.
* `system_disk_size` - (Required, Int, ForceNew) Size of the system disk. unit is GiB. If modified, the ZEC instance may force stop.
* `disable_qga_agent` - (Optional, Bool) Indicate whether to disable QEMU Guest Agent (QGA). QGA is enabled by default. Changing `disable_qga_agent` will cause the ZEC instance reset.
* `enable_ip_forwarding` - (Optional, Bool) Indicate whether to enable IP forwarding. IP forwarding is disabled by default.
* `force_delete` - (Optional, Bool) Indicate whether to force delete the ZEC instance. Default is `true`. If set true, the ZEC instance will be permanently deleted instead of being moved into the recycle bin.
* `instance_name` - (Optional, String) The name of the ZEC instance. The minimum length of instance name is `2`. The max length of instance_name is 63, and default value is `Terraform-ZEC-Instance`.
* `key_id` - (Optional, String) The key pair id to use for the ZEC instance. Changing `key_id` will cause the ZEC instance reset.
* `password` - (Optional, String) Password for the ZEC instance.The password must be 8-16 characters, including letters, numbers, and special characters `~!@$^*-_=+|;:,.?`.
* `resource_group_id` - (Optional, String) The resource group id the ZEC instance belongs to, default to Default Resource Group.
* `running_flag` - (Optional, Bool) Set instance to running or stop. Default value is true, the instance will shutdown when this flag is false.
* `system_disk_category` - (Optional, String, ForceNew) Category of the system disk. Valid values: `Standard NVMe SSD`, `Basic NVMe SSD`, Default is `Standard NVMe SSD`.
* `time_zone` - (Optional, String) Time zone of instance. such as `America/Los_Angeles`. Default is `Asia/Shanghai`. Changing `time_zone` will cause the ZEC instance reset.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `cpu` - The number of CPU cores of the ZEC instance.
* `create_time` - Create time of the ZEC instance.
* `image_name` - The image name to use for the ZEC instance.
* `instance_status` - Current status of the ZEC instance.
* `memory` - Memory capacity of the ZEC instance, unit in GiB.
* `private_ip_addresses` - Private Ip addresses of the ZEC instance.
* `public_ip_addresses` - Public Ip addresses of the ZEC instance.
* `resource_group_name` - The resource group name the ZEC instance belongs to, default to Default Resource Group.
* `system_disk_id` - ID of the system disk.


