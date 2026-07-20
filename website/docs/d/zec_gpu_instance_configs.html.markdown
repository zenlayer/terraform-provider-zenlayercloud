---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_gpu_instance_configs"
sidebar_current: "docs-zenlayercloud-datasource-zec_gpu_instance_configs"
description: |-
  Use this data source to query available GPU instance type configurations in Zenlayer ZEC zones.
---

# zenlayercloud_zec_gpu_instance_configs

Use this data source to query available GPU instance type configurations in Zenlayer ZEC zones.

## Example Usage

Query all GPU instance configs

```hcl
data "zenlayercloud_zec_gpu_instance_configs" "all" {
}
```

Query GPU configs in a specific zone

```hcl
data "zenlayercloud_zec_gpu_instance_configs" "by_zone" {
  zone_id = "na-us-la-2a"
}
```

## Argument Reference

The following arguments are supported:

* `instance_type` - (Optional, String) GPU instance type to filter, e.g. `z3a.g.C49.c8m32.1`. Returns all types if not specified.
* `result_output_file` - (Optional, String) Used to save results.
* `zone_id` - (Optional, String) Zone ID to filter GPU instance configs, e.g. `na-us-la-2a`. Returns all zones if not specified.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `data_set` - List of GPU instance type quota items. Each element contains the following attributes:
   * `bps` - Maximum bandwidth per NIC in bits per second.
   * `cpu_count` - Number of vCPUs.
   * `gpu_amount` - Number of GPU cards.
   * `instance_type_name` - Human-readable description of the GPU instance type.
   * `instance_type` - GPU instance type ID, e.g. `z3a.g.C49.c8m32.1`. Pass this value to `instance_type` when creating or modifying an instance.
   * `inventory_capacity` - Inventory capacity level for this GPU instance type.
   * `memory` - Memory size in GiB.
   * `pps` - Maximum packets per second per NIC.
   * `zone_id` - Zone ID, e.g. `na-us-la-2a`.


