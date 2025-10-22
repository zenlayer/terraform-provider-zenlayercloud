Use this data source to query vpc route entries.

Example Usage

Query all vpc route entries
```hcl

data "zenlayercloud_zec_vpc_routes" "all" {
}
```

Query vpc route entries by vpc id

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
	vpc_id = "vpc-xxxxxx"
}
```

Query vpc route entries by name regex

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
	name_regex = "^vpc-"
}
```

Query vpc route entries by destination cidr block

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
	destination_cidr_block = "10.0.0.0/16"
}
```

Query vpc route entries by route type

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
	route_type = "RouteTypeStatic"
}
```


Query vpc route entries by ip version

```hcl
data "zenlayercloud_zec_vpc_routes" "foo" {
	ip_version = "IPv4"
}
```
