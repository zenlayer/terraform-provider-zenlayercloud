---
subcategory: "Zenlayer Load Balancing(ZLB)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zlb_backend"
sidebar_current: "docs-zenlayercloud-resource-zlb_backend"
description: |-
  Provide a resource to create a backend instances for ZLB listener.
---

# zenlayercloud_zlb_backend

Provide a resource to create a backend instances for ZLB listener.

~> **NOTE:** The current resource is used to manage all backend servers under one listener, and it is not allowed for the same listener to use multiple current resources to manage them at the same time.

## Example Usage

Prepare a ZLB instance, listener and ZEC instances

```hcl
# VPC
resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}
resource "zenlayercloud_zec_vpc_security_group_attachment" "foo" {
  vpc_id            = zenlayercloud_zec_vpc.foo.id
  security_group_id = "1304682049596034008"
}

# ZLB instance & listener
resource "zenlayercloud_zlb_instance" "zlb" {
  region_id = var.region
  vpc_id    = zenlayercloud_zec_vpc.foo.id
  zlb_name  = "example-5"
}

resource "zenlayercloud_zlb_listener" "tcp_listener" {
  zlb_id               = zenlayercloud_zlb_instance.zlb.id
  listener_name        = "tcp-listener"
  protocol             = "TCP"
  health_check_enabled = true
  port                 = 8080
  scheduler            = "mh"
  kind                 = "FNAT"
  health_check_type    = "TCP"
}


# Subnet & Instance
resource "zenlayercloud_zec_subnet" "ipv4" {
  vpc_id     = zenlayercloud_zec_vpc.foo.id
  region_id  = var.region
  name       = "example-instance-subnet"
  cidr_block = "10.0.0.0/24"
}

data "zenlayercloud_key_pairs" "all" {
}

data "zenlayercloud_zec_images" "ubuntu" {
  availability_zone = var.availability_zone
  category          = "Ubuntu"
}

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

Create a backend instance

```hcl
resource "zenlayercloud_zlb_backend" "backend" {
  zlb_id      = zenlayercloud_zlb_instance.zlb.id
  listener_id = split(":", zenlayercloud_zlb_listener.tcp_listener.id)[1]
  backends {
    instance_id        = zenlayercloud_zec_instance.instance.id
    private_ip_address = zenlayercloud_zec_instance.instance.private_ip_addresses[0]
  }
}
```

# Import

ZLB backends can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zlb_backend.backends zlb-id : listener-id
```

## Argument Reference

The following arguments are supported:

* `backends` - (Required, Set) List of backend servers.
* `listener_id` - (Required, String, ForceNew) ID of the listener.
* `zlb_id` - (Required, String, ForceNew) ID of the load balancer instance.

The `backends` object supports the following:

* `private_ip_address` - (Required, String) Private IP address of the network interface attached to the instance.
* `instance_id` - (Optional, String) ID of the backend server. The added instance must belong to the VPC associated with lb.
* `port` - (Optional, Int) Target port for request forwarding and health checks. If left empty, it will follow the listener's port configuration. Valid values: `1` to `65535`.
* `weight` - (Optional, Int) Forwarding weight of the backend server. Valid value ranges: (0~65535). Default to 100. Weight of 0 means the server will not accept new requests.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.



