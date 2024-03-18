---
subcategory: "Bare Metal Cloud(BMC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_bmc_ddos_ips"
sidebar_current: "docs-zenlayercloud-datasource-bmc_ddos_ips"
description: |-
  Use this data source to query DDoS IP instances.
---

# zenlayercloud_bmc_ddos_ips

Use this data source to query DDoS IP instances.

## Example Usage

```hcl
data "zenlayercloud_bmc_ddos_ips" "foo" {
  availability_zone = "SEL-A"
}
```

## Argument Reference

The following arguments are supported:

* `associated_instance_id` - (Optional, String) The ID of instance to bind with DDoS IPs to be queried.
* `availability_zone` - (Optional, String) The ID of zone that the DDoS IPs locates at.
* `ip_ids` - (Optional, Set: [`String`]) IDs of the DDoS IP to be queried.
* `ip_status` - (Optional, String) The status of elastic ip to be queried.
* `public_ip` - (Optional, String) The address of elastic ip to be queried.
* `resource_group_id` - (Optional, String) The ID of resource group grouped instances to be queried.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ip_list` - An information list of DDoS IP. Each element contains the following attributes:
   * `availability_zone` - The ID of zone that the DDoS IP locates at.
   * `create_time` - Create time of the DDoS IP.
   * `expired_time` - Expired time of the DDoS IP.
   * `instance_id` - The instance id to bind with the DDoS IP.
   * `instance_name` - The instance name to bind with the DDoS IP.
   * `ip_charge_type` - The charge type of DDoS IP.
   * `ip_id` - ID  of the DDoS IP.
   * `ip_status` - Current status of the DDoS IP.
   * `public_ip` - The elastic ip address.
   * `resource_group_id` - The ID of resource group grouped instances to be queried.
   * `resource_group_name` - The name of resource group grouped instances to be queried.


