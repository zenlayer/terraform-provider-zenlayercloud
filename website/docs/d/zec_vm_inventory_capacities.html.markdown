---
subcategory: "Zenlayer Elastic Compute(ZEC)"
layout: "zenlayercloud"
page_title: "ZenlayerCloud: zenlayercloud_zec_vm_inventory_capacities"
sidebar_current: "docs-zenlayercloud-datasource-zec_vm_inventory_capacities"
description: |-
  Use this data source to query ZEC VM inventory capacity levels by node and instance type.
---

# zenlayercloud_zec_vm_inventory_capacities

Use this data source to query ZEC VM inventory capacity levels by node and instance type.

Capacity is expressed as one of four levels based on the total number of sellable CPU cores across all instance types:

| Level | Threshold |
|-------|-----------|
| `LIMITED` | < 1000 cores |
| `NORMAL` | 1000-2000 cores |
| `SUFFICIENT` | 2000-5000 cores |
| `ABUNDANT` | >= 5000 cores |

## Example Usage

Query all nodes

```hcl
data "zenlayercloud_zec_vm_inventory_capacities" "all" {
}
```

Query specific nodes by region ID

```hcl
data "zenlayercloud_zec_vm_inventory_capacities" "filtered" {
  region_ids = ["asia-north-1", "eu-west-1"]
}
```

## Argument Reference

The following arguments are supported:

* `region_ids` - (Optional, Set: [`String`]) Node IDs to query, e.g. `asia-north-1`. Returns all nodes if not specified.
* `result_output_file` - (Optional, String) Used to save results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `data_set` - Inventory capacity list per node. Each element contains the following attributes:
   * `capacity` - Overall inventory capacity level of the node. One of `LIMITED` (< 1000 cores), `NORMAL` (1000-2000 cores), `SUFFICIENT` (2000-5000 cores), `ABUNDANT` (>= 5000 cores).
   * `instance_types` - Per-instance-type capacity breakdown. Entries with zero inventory are excluded.
      * `capacity` - Inventory capacity level for this instance type. One of `LIMITED` (< 1000 cores), `NORMAL` (1000-2000 cores), `SUFFICIENT` (2000-5000 cores), `ABUNDANT` (>= 5000 cores).
      * `gpu_spec` - GPU model, e.g. `z4a.g.C49`. Only present for GPU instances.
      * `instance_type` - CPU instance type, e.g. `z2a`, `z2i`, `z4a`.
   * `region_id` - Node ID, e.g. `asia-north-1`.


