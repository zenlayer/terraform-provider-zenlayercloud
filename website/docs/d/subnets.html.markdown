---
subcategory: "Zenlayer Virtual Machine(VM)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_subnets"
sidebar_current: "docs-zenlayercloud-datasource-subnets"
description: |-
  Use this data source to query subnets information.
---

# zenlayercloud_subnets

Use this data source to query subnets information.

## Example Usage

```hcl
variable "availability_zone" {
  default = "SEL-A"
}

resource "zenlayercloud_subnet" "foo" {
  availability_zone = var.availability_zone
  name              = "subnet_test"
  cidr_block        = "10.0.0.0/24"
}

# filter by subnet id
data "zenlayercloud_subnets" "id_subnets" {
  subnet_id = zenlayercloud_subnet.foo.id
}

# filter by subnet name
data "zenlayercloud_subnets" "name_subnets" {
  subnet_name = zenlayercloud_subnet.foo.name
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional, String) Zone of the subnet to be queried.
* `cidr_block` - (Optional, String) Filter subnet with this CIDR.
* `result_output_file` - (Optional, String) Used to save results.
* `subnet_id` - (Optional, String) ID of the subnet to be queried.
* `subnet_name` - (Optional, String) Name of the subnet to be queried.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `subnet_list` - An information list of subnet. Each element contains the following attributes:
  * `availability_zone` - The availability zone of the subnet.
  * `cidr_block` - A network address block of the subnet.
  * `create_time` - Creation time of the subnet.
  * `subnet_id` - ID of the subnet.
  * `subnet_name` - Name of the subnet.
  * `subnet_status` - Status of the subnet.


