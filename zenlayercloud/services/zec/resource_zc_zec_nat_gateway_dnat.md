Provides a resource to create a NAT Gateway DNAT forwarding entry.

Example Usage

Prepare a NAT gateway
```hcl

variable "region_shanghai" {
  default = "asia-east-1"
}


resource "zenlayercloud_zec_nat_gateway" "foo" {
  region_id         = var.region_shanghai
  name              = "test-nat"
  vpc_id            = "<vpc_id>"
  security_group_id = "<security_group_id>"
  subnet_ids = ["<subnet_id>"]
}

```


Create a DNat forwarding entry
```hcl

resource "zenlayercloud_zec_nat_gateway_dnat" "foo" {
	nat_gateway_id = zenlayercloud_zec_nat_gateway.foo.id
	source_cidr_block = ["10.0.0.0/8"]
	eip_ids = ["192.168.1.1"]
}
```

Import

Dnat entry can be imported using the id, the id format must be '{nat_gateway_id}:{dnat_id}'

```
$ terraform import zenlayercloud_zec_nat_gateway_dnat.foo nat-gateway-id:dnat-id
```
