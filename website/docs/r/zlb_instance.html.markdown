---
subcategory: "Zenlayer Load Balancing(ZLB)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zlb_instance"
sidebar_current: "docs-zenlayercloud-resource-zlb_instance"
description: |-
  Provide a resource to create a ZLB instance.
---

# zenlayercloud_zlb_instance

Provide a resource to create a ZLB instance.

## Example Usage

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_vpc" "foo" {
  name        = "example"
  cidr_block  = "10.0.0.0/16"
  enable_ipv6 = true
}

resource "zenlayercloud_zlb_instance" "zlb" {
  region_id = var.region
  vpc_id    = zenlayercloud_zec_vpc.foo.id
  zlb_name  = "example-5"
}
```

# Import

ZLB instance can be imported, e.g.

```hcl
$ terraform import zenlayercloud_zlb_instance.zlb zlb-id
```

## Argument Reference

The following arguments are supported:

* `region_id` - (Required, String, ForceNew) The ID of region that the load balancer instance locates at.
* `vpc_id` - (Required, String, ForceNew) The ID of VPC that the load balancer instance belongs to.
* `resource_group_id` - (Optional, String) The resource group id the load balancer belongs to, default to Default Resource Group.
* `zlb_name` - (Optional, String) The name of the load balancer instance.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the resource.
* `create_time` - Create time of the load balancer instance.
* `private_ip_addresses` - Private virtual Ipv4 addresses of the load balancer instance.
* `public_ip_addresses` - Public IPv4 addresses(EIP) of the load balancer instance.
* `resource_group_name` - The resource group name the load balancer belongs to.
* `zlb_status` - Status of the load balancer instance.


