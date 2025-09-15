---
subcategory: "Zenlayer Load Balancing(ZLB)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zlb_instances"
sidebar_current: "docs-zenlayercloud-datasource-zlb_instances"
description: |-
  Use this data source to query ZLB instances.
---

# zenlayercloud_zlb_instances

Use this data source to query ZLB instances.

## Example Usage

Query all ZLB instances

```hcl
data "zenlayercloud_zlb_instances" "all" {
}
```

Query ZLB instances by ids

```hcl
data "zenlayercloud_zlb_instances" "foo" {
  ids = ["<zlbId>"]
}
```

Query ZLB instances by region id

```hcl
variable "region" {
  default = "asia-east-1"
}

data "zenlayercloud_zlb_instances" "foo" {
  region_id = var.region
}
```

Query ZLB instances by vpc id

```hcl
data "zenlayercloud_zlb_instances" "foo" {
  vpc_id = "<vpcId>"
}
```

Query ZLB instances by name regex

```hcl
data "zenlayercloud_zlb_instances" "foo" {
  name_regex = "Web*"
}
```

## Argument Reference

The following arguments are supported:

* `ids` - (Optional, Set: [`String`]) IDs of the load balancer instances to be queried.
* `name_regex` - (Optional, String) A regex string to filter results by  load balancer instance name.
* `region_id` - (Optional, String) The ID of region that the load balancer instances locates at.
* `resource_group_id` - (Optional, String) The ID of resource group that the load balancer instance grouped by.
* `result_output_file` - (Optional, String) Used to save results.
* `vpc_id` - (Optional, String) ID of the VPC to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `zlbs` - An information list of instances. Each element contains the following attributes:
   * `create_time` - Create time of the load balancer instance.
   * `private_ip_addresses` - Private virtual Ipv4 addresses of the load balancer instance.
   * `public_ip_addresses` - Public IPv4 addresses(EIP) of the load balancer instance.
   * `region_id` - The ID of region that the load balancer instance locates at.
   * `resource_group_id` - The ID of resource group grouped load balancer instance to be queried.
   * `resource_group_name` - The name of resource group that the load balancer instance belongs to.
   * `status` - Current status of the load balancer instance.
   * `vpc_id` - VPC ID to which the load balance belongs.
   * `zlb_id` - ID of the load balancer instances.
   * `zlb_name` - The name of the load balancer instance.


