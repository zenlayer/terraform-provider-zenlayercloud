Use this data source to query images.

Example Usage

```hcl
variable "availability_zone" {
	default = "asia-east-1a"
}

data "zenlayercloud_zec_images" "foo" {
	availability_zone =  var.availability_zone
}
```
