Use this data source to query ZEC instances.

Example Usage

Query all instances
```hcl
data "zenlayercloud_zec_instances" "foo" {
}
```

Query zec instances by ids
```hcl
data "zenlayercloud_zec_instances" "foo" {
	ids = ["<instanceId>"]
}
```

Query zec instances by availability zone
```hcl
data "zenlayercloud_zec_instances" "foo" {
	availability_zone = "asia-southeast-1a"
}
```

Query zec instances by name regex 
```hcl
data "zenlayercloud_zec_instances" "foo" {
	name_regex = "test*"
}
```

Query zec instances by image id
```hcl
data "zenlayercloud_zec_instances" "foo" {
	image_id = "<imageId>"
}	
```

Query zec instances by IPv4 address (including private & public IPv4)
```hcl
data "zenlayercloud_zec_instances" "foo" {
	ipv4_address = "10.0.0.2"
}	
```
