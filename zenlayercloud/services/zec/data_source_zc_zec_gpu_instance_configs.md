Use this data source to query available GPU instance type configurations in Zenlayer ZEC zones.

Example Usage

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
