Provide a resource to create CIDR block.

Example Usage

```hcl
variable "region" {
  default = "asia-east-1"
}

resource "zenlayercloud_zec_cidr" "test" {
  region_id    = var.region
  netmask      = 27
  network_type = "BGPLine"
  tags = {
    "group"  = "test"
  }
}
```

# Import

CIDR block can be imported, e.g.

```
$ terraform import zenlayercloud_zec_cidr.test cidr-block-id
```
