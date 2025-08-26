---
subcategory: "Zenlayer Load Balancing(ZLB)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zlb_regions"
sidebar_current: "docs-zenlayercloud-datasource-zlb_regions"
description: |-
  Use this data source to query available regions for load balancer.
---

# zenlayercloud_zlb_regions

Use this data source to query available regions for load balancer.

## Example Usage

Query all load balancer regions

```hcl
data "zenlayercloud_zlb_regions" "all" {
}
```

Query load balancer regions by city code

```hcl
data "zenlayercloud_zlb_regions" "foo" {
  city_code = "SEL" s
}
```

## Argument Reference

The following arguments are supported:

* `city_code` - (Optional, String) The code of the city where the region is located to be queried.
* `region_id` - (Optional, String) The ID of region that the load balancer locates at.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `regions` - An information list of instances. Each element contains the following attributes:
   * `city_code` - The code of the city where the region is located. such as `SHA`.
   * `city_name` - The name of the city where the region is located. such as `Shanghai`.
   * `region_id` - The ID of region that support for load balancer. such as `asia-east-1`.


