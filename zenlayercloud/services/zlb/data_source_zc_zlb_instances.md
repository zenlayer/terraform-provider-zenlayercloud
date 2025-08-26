Use this data source to query ZLB instances.

Example Usage

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

