Use this data source to query ZEC VM inventory capacity levels by node and instance type.

Capacity is expressed as one of four levels based on the total number of sellable CPU cores across all instance types:

| Level | Threshold |
|-------|-----------|
| `LIMITED` | < 1000 cores |
| `NORMAL` | 1000-2000 cores |
| `SUFFICIENT` | 2000-5000 cores |
| `ABUNDANT` | >= 5000 cores |

Example Usage

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
