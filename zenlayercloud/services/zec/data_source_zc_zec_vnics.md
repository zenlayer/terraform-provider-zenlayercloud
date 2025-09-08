Use this data source to query vNIC information.

Example Usage

Query all vNICs
```hcl
data "zenlayercloud_zec_vnics" "foo" {
}
```

Query vNICs by ids
```hcl
data "zenlayercloud_zec_vnics" "foo" {
	ids = ["<vnicId>"]
}
```

Query vNICs by region id
```hcl
data "zenlayercloud_zec_vnics" "foo" {
	region_id = "asia-southeast-1"
}
```

Query vNICs by name regex
```hcl
data "zenlayercloud_zec_vnics" "foo" {
	name_regex = "test*"
}
```

Query vNICs by subnet id
```hcl
data "zenlayercloud_zec_vnics" "foo" {
	subnet_id = "<subnetId>"
}	
```

Query vNICs by vpc id
```hcl
data "zenlayercloud_zec_vnics" "foo" {
	vpc_id = "<vpcId>"
}	
```

Query vNICs by associated ZEC instance id
```hcl
data "zenlayercloud_zec_vnics" "foo" {
	instance_id = "<instanceId>"
}	
```
